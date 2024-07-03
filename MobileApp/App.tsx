import React, {useCallback, useEffect, useRef, useState} from 'react';
import {
  SafeAreaView,
  ScrollView,
  StatusBar,
  StyleSheet,
  Text,
  View,
  TouchableOpacity,
  Modal,
  NativeEventEmitter,
  useColorScheme,
  TextInput,
  FlatList,
} from 'react-native';

import {Buffer} from 'buffer';
global.Buffer = Buffer; // Ensure Buffer is available globally

import Icon from 'react-native-vector-icons/MaterialCommunityIcons';
import {NativeModules} from 'react-native';
const {PeerModule} = NativeModules;

import {generateOrder, updateOrder} from './utils';
import OrderCard from './orderCard';
import {
  addGlobalEvents,
  areObjectsEqual,
  generateFullOrdersFromCache,
  getLastEvent,
} from './globalcache';

let isDarkMode;
function App() {
  isDarkMode = useColorScheme() === 'dark';
  const [orders, setOrders] = useState({});
  const [modalVisible, setModalVisible] = useState(false);
  const [peers, setPeers] = useState([]);
  const [peerId, setPeerId] = useState('');
  const emitterRef = useRef(new NativeEventEmitter(PeerModule));

  const [searchTerm, setSearchTerm] = useState('');

  const backgroundStyle = {
    backgroundColor: isDarkMode ? '#333' : '#FFF',
  };

  const textColor = {
    color: isDarkMode ? '#FFF' : '#333',
  };

  const peerBlockStyle = {
    padding: 10,
  };

  const peerCardStyle = {
    backgroundColor: isDarkMode ? '#777' : '#eee',
    padding: 10,
    borderRadius: 10,
    marginBottom: 10,
    shadowColor: '#000',
    shadowOffset: {
      width: 0,
      height: 2,
    },
    shadowOpacity: 0.25,
    shadowRadius: 3.84,
    elevation: 5,
  };

  const handleReceivedData = useCallback(
    async data => {
      let ordersData;
      try {
        ordersData = JSON.parse(data.message);
        // console.log('Parsed ordersData:', ordersData);
      } catch (error) {
        console.error('Failed to parse data.message:', error);
        return;
      }

      if (!Array.isArray(ordersData)) {
        console.error('Parsed data.message is not an array:', ordersData);
        return;
      }
      const newEvents = {};
      for (const childData of ordersData) {
        let orderData;
        try {
          orderData = JSON.parse(childData);
        } catch (error) {
          console.error('Failed to parse childData:', error);
          return;
        }

        if (!orderData.orderId || !Array.isArray(orderData.events)) {
          console.error('Invalid orderData structure:', orderData);
          return;
        }

        if (!newEvents[orderData.orderId]) {
          newEvents[orderData.orderId] = [];
        }
        newEvents[orderData.orderId].push(...orderData.events);
      }

      // Add global events in bulk
      Object.keys(newEvents).forEach(orderId => {
        try {
          addGlobalEvents(orderId, newEvents[orderId]);
        } catch (error) {
          console.error(
            `Failed to add global events for orderId ${orderId}:`,
            error,
          );
        }
      });

      let newOrders;
      try {
        newOrders = generateFullOrdersFromCache(orders);
      } catch (error) {
        console.error('Failed to generate full orders from cache:', error);
        return;
      }
      setOrders(prevOrders => {
        if (!areObjectsEqual(newOrders, prevOrders)) {
          console.log('Setting new state');
          const out = {};
          for (const order of newOrders) {
            out[order.id] = order;
          }
          return out;
        }
        console.log('State not changed');
        return prevOrders;
      });
    },
    [orders],
  );

  useEffect(() => {
    const peersListener = emitterRef.current.addListener('PEERS', data => {
      setPeers(data.message.split(','));
    });
    const peerIdListener = emitterRef.current.addListener('PEER_ID', data => {
      setPeerId(data.message);
    });

    PeerModule.start();
    return () => {
      peersListener.remove();
      peerIdListener.remove();
    };
  }, []);

  useEffect(() => {
    const orderListener = emitterRef.current.addListener(
      'P2P',
      handleReceivedData,
    );

    return () => {
      orderListener.remove();
    };
  }, [handleReceivedData]);

  const createOrder = index => {
    const newOrderEvents = generateOrder(index);
    const message = Buffer.from(
      JSON.stringify({
        orderId: newOrderEvents[0].orderId,
        events: newOrderEvents,
      }),
    ).toString('hex');
    PeerModule.sendMessage(message);
  };

  // const getLogs = () => {
  //   try {
  //     const logs = PeerModule.getLogs();
  //     const logsArray = JSON.parse(logs) as Array<string>;
  //     return logsArray.join('\n');
  //   } catch (e) {
  //     return 'error';
  //   }
  // };

  const updateOrderEvent = useCallback((orderId: string) => {
    const updatedEvents = updateOrder(orderId, getLastEvent(orderId));
    const message = Buffer.from(
      JSON.stringify({
        orderId: orderId,
        events: updatedEvents,
      }),
    ).toString('hex');
    console.log('Sending message to native');
    PeerModule.sendMessage(message);
  }, []);

  const renderItem = ({item, index}) => (
    <OrderCard
      key={item.id}
      onUpdateOrder={updateOrderEvent}
      number={index}
      orderId={item.id}
      totalPrice={item.totalPrice}
      status={item.status}
    />
  );

  return (
    <SafeAreaView style={[styles.container, backgroundStyle]}>
      <StatusBar
        barStyle={isDarkMode ? 'light-content' : 'dark-content'}
        backgroundColor={backgroundStyle.backgroundColor}
      />
      <FlatList
        contentInsetAdjustmentBehavior="automatic"
        style={backgroundStyle}
        data={Object.values(orders)}
        renderItem={renderItem}
        keyExtractor={(item, index) => `${item.id}-${index}`}
      />
      <View style={styles.buttonContainer}>
        {Array.from({length: 3}).map((_, index) => (
          <TouchableOpacity
            key={index}
            style={styles.iconButton}
            onPress={() => createOrder(index)}>
            <Icon name="plus-box" size={20} color="#fff" />
            <Text style={styles.iconText}>Create Order {index + 1}</Text>
          </TouchableOpacity>
        ))}
        <TouchableOpacity
          style={styles.iconButton}
          onPress={() => setModalVisible(true)}>
          <Icon name="account-multiple" size={20} color="#fff" />
          <Text style={styles.iconText}>Show Peers</Text>
        </TouchableOpacity>
      </View>
      <Modal
        animationType="slide"
        transparent={true}
        visible={modalVisible}
        onRequestClose={() => setModalVisible(!modalVisible)}>
        <View style={styles.centeredView}>
          <View
            style={[
              styles.modalView,
              {backgroundColor: isDarkMode ? '#555' : '#fff'},
            ]}>
            <Text style={[styles.modalText, textColor]}>Connected Peers:</Text>
            <TextInput // New TextInput for the search input
              style={styles.searchInput}
              value={searchTerm}
              onChangeText={setSearchTerm}
              placeholder="Search for a peer id..."
            />
            <ScrollView style={peerBlockStyle}>
              {peers
                .filter(peer => peer.toString().includes(searchTerm)) // Filter the peers based on the search term
                .map((peer, index) => {
                  return (
                    <View key={index} style={peerCardStyle}>
                      <Text style={[styles.peerText, textColor]}>
                        {peer} {peer === peerId ? '(You)' : ''}
                      </Text>
                    </View>
                  );
                })}
            </ScrollView>
            <Text>Your peer id = {peerId}</Text>
            <TouchableOpacity
              style={styles.buttonClose}
              onPress={() => setModalVisible(!modalVisible)}>
              <Text style={styles.textStyle}>Hide Modal</Text>
            </TouchableOpacity>
          </View>
        </View>
      </Modal>
      {/* <Modal
        animationType="slide"
        transparent={true}
        visible={showLogModale}
        onRequestClose={() => setShowLogModal(!showLogModale)}>
        <View style={styles.centeredView}>
          <View
            style={[
              styles.modalView,
              // eslint-disable-next-line react-native/no-inline-styles
              {backgroundColor: isDarkMode ? '#555' : '#fff'},
            ]}>
            <ScrollView>
              <Text selectable>{getLogs()}</Text>
              <TouchableOpacity
                style={styles.buttonClose}
                onPress={() => setShowLogModal(!showLogModale)}>
                <Text style={styles.textStyle}>Hide Modal</Text>
              </TouchableOpacity>
            </ScrollView>
          </View>
        </View>
      </Modal> */}
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 20,
  },
  buttonContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'space-evenly',
    marginTop: 20,
  },
  iconButton: {
    flexDirection: 'row',
    backgroundColor: '#007bff',
    padding: 10,
    borderRadius: 20,
    alignItems: 'center',
    justifyContent: 'center',
    margin: 5,
  },
  iconText: {
    color: '#fff',
    marginLeft: 5,
  },
  centeredView: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    marginTop: 22,
  },
  modalView: {
    margin: 20,
    borderRadius: 20,
    padding: 35,
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: {
      width: 0,
      height: 2,
    },
    shadowOpacity: 0.25,
    shadowRadius: 4,
    elevation: 5,
  },
  modalText: {
    marginBottom: 15,
    textAlign: 'center',
  },
  buttonClose: {
    backgroundColor: '#2196F3',
    borderRadius: 20,
    padding: 10,
    elevation: 2,
  },
  textStyle: {
    color: 'white',
    fontWeight: 'bold',
    textAlign: 'center',
  },
  peersContainer: {
    maxHeight: 200, // Adjust this value as needed
  },
  peerText: {
    fontSize: 16,
  },
  searchInput: {
    // New style for the search input
    height: 40,
    borderColor: 'gray',
    borderWidth: 1,
    borderRadius: 10,
    padding: 10,
    marginBottom: 10,
  },
});

export default App;
