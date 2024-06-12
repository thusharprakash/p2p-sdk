import React, {useEffect, useState} from 'react';
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
} from 'react-native';

import {Buffer} from 'buffer';
global.Buffer = Buffer; // Ensure Buffer is available globally

import Icon from 'react-native-vector-icons/MaterialCommunityIcons';
import {NativeModules} from 'react-native';
const {PeerModule} = NativeModules;

import {generateOrder, updateOrder} from './utils';
import {computeOrderState} from '@oolio-group/order-helper';
import OrderCard from './orderCard';
let isDarkMode;
function App() {
  isDarkMode = useColorScheme() === 'dark';
  const [orders, setOrders] = useState({});
  const [lastProcessedEventIds, setLastProcessedEventIds] = useState(new Map());
  const [modalVisible, setModalVisible] = useState(false);
  const [peers, setPeers] = useState([]);
  const [peerId, setPeerId] = useState('');

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
  useEffect(() => {
    const emitter = new NativeEventEmitter(PeerModule);
    const orderListener = emitter.addListener('P2P', handleReceivedData);
    const peersListener = emitter.addListener('PEERS', data => {
      setPeers(data.message.split(','));
    });
    const peerIdListener = emitter.addListener('PEER_ID', data => {
      setPeerId(data.message);
    });

    PeerModule.start();

    return () => {
      orderListener.remove();
      peersListener.remove();
      peerIdListener.remove();
    };
  }, []);

  const handleReceivedData = data => {
    const message = Buffer.from(data.message, 'hex').toString();
    const orderData = JSON.parse(message);
    console.log('Received order data:', orderData);

    setOrders(prevOrders => {
      const prevOrder = prevOrders[orderData.orderId];
      const newOrder = computeOrderState(
        orderData.events,
        prevOrder || undefined,
      );
      console.log('Previous order:', prevOrder);
      console.log('New order:', newOrder);

      return {...prevOrders, [orderData.orderId]: newOrder};
    });

    setLastProcessedEventIds(prevEventIds => {
      const lastEventId = orderData.events[orderData.events.length - 1].id;
      return new Map(prevEventIds).set(orderData.orderId, lastEventId);
    });
  };

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

  const updateOrderEvent = orderId => {
    const lastEventId = lastProcessedEventIds.get(orderId);
    const updatedEvents = updateOrder(orderId, lastEventId);
    const message = Buffer.from(
      JSON.stringify({
        orderId: orderId,
        events: updatedEvents,
      }),
    ).toString('hex');
    PeerModule.sendMessage(message);
  };

  return (
    <SafeAreaView style={[styles.container, backgroundStyle]}>
      <StatusBar
        barStyle={isDarkMode ? 'light-content' : 'dark-content'}
        backgroundColor={backgroundStyle.backgroundColor}
      />
      <ScrollView
        contentInsetAdjustmentBehavior="automatic"
        style={backgroundStyle}>
        {Object.entries(orders).map(([id, order]) => (
          <OrderCard
            key={id}
            order={order}
            onUpdateOrder={() => updateOrderEvent(id)}
            lastProcessedEventIds={lastProcessedEventIds}
          />
        ))}
      </ScrollView>
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
