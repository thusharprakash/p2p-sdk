import React, {useState} from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  useColorScheme,
  LayoutAnimation,
  UIManager,
  Platform,
} from 'react-native';
import Icon from 'react-native-vector-icons/MaterialCommunityIcons';

// Enable LayoutAnimation on Android
if (Platform.OS === 'android') {
  UIManager.setLayoutAnimationEnabledExperimental &&
    UIManager.setLayoutAnimationEnabledExperimental(true);
}

const OrderCard = ({
  order,
  onUpdateOrder,
  lastProcessedEventIds,
  number,
}: {
  order: any;
  onUpdateOrder: any;
  lastProcessedEventIds: any;
  number: number;
}) => {
  const [collapsed, setCollapsed] = useState(true);
  const isDarkMode = useColorScheme() === 'dark';
  const lastEventId = lastProcessedEventIds.get(order.id);
  const orderId = order.id;

  const toggleCollapse = () => {
    LayoutAnimation.configureNext(LayoutAnimation.Presets.easeInEaseOut);
    setCollapsed(!collapsed);
  };

  const cardStyle = {
    backgroundColor: isDarkMode ? '#444' : '#fff',
    shadowColor: isDarkMode ? '#fff' : '#000',
  };

  const textStyle = {
    color: isDarkMode ? '#fff' : '#000',
  };

  return (
    <View style={[styles.card, cardStyle]}>
      <View style={styles.header}>
        <Text style={[styles.title, textStyle]}>
          {number + 1}.Order ID: {orderId}
        </Text>
        <TouchableOpacity onPress={toggleCollapse}>
          <Icon
            name={collapsed ? 'chevron-down' : 'chevron-up'}
            size={24}
            color={textStyle.color}
          />
        </TouchableOpacity>
      </View>
      <Text style={[styles.text, textStyle]}>
        Order Total: ${order.totalPrice}
      </Text>
      <Text style={[styles.text, textStyle]}>Status: {order.status}</Text>
      {!collapsed && (
        <View style={styles.itemsContainer}>
          {order.orderItems?.map((item: any, index: any) => (
            <View key={index} style={[styles.itemCard, cardStyle]}>
              <Text style={[styles.itemText, textStyle]}>
                {item.product?.name || 'Unknown Product'}
              </Text>
              <Text style={[styles.itemText, textStyle]}>
                Quantity: {item.quantity}, Price: ${item.unitPrice}
              </Text>
            </View>
          ))}
        </View>
      )}
      <TouchableOpacity
        style={styles.updateButton}
        onPress={() => onUpdateOrder(orderId, lastEventId)}>
        <Text style={styles.buttonText}>Update Order</Text>
      </TouchableOpacity>
    </View>
  );
};

const styles = StyleSheet.create({
  card: {
    padding: 20,
    margin: 10,
    borderRadius: 10,
    shadowOffset: {width: 0, height: 2},
    shadowOpacity: 0.25,
    shadowRadius: 3.84,
    elevation: 5,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  title: {
    fontSize: 18,
    fontWeight: 'bold',
    marginBottom: 10,
  },
  text: {
    fontSize: 16,
    marginBottom: 5,
  },
  itemsContainer: {
    marginTop: 10,
  },
  itemCard: {
    padding: 10,
    marginVertical: 5,
    borderRadius: 8,
    shadowOffset: {width: 0, height: 1},
    shadowOpacity: 0.2,
    shadowRadius: 2.22,
    elevation: 3,
  },
  itemText: {
    fontSize: 14,
  },
  updateButton: {
    marginTop: 15,
    backgroundColor: '#007bff',
    padding: 10,
    borderRadius: 10,
    alignItems: 'center',
  },
  buttonText: {
    color: '#fff',
    fontWeight: 'bold',
  },
});

export default OrderCard;
