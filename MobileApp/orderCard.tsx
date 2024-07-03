import React from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  useColorScheme,
  UIManager,
  Platform,
} from 'react-native';

// Enable LayoutAnimation on Android
if (Platform.OS === 'android') {
  UIManager.setLayoutAnimationEnabledExperimental &&
    UIManager.setLayoutAnimationEnabledExperimental(true);
}

const OrderCard = ({
  orderId,
  totalPrice,
  status,
  onUpdateOrder,
  number,
}: {
  orderId: string;
  totalPrice: number;
  status: string;
  onUpdateOrder: any;
  number: number;
}) => {
  const isDarkMode = useColorScheme() === 'dark';

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
      </View>
      <Text style={[styles.text, textStyle]}>Order Total: ${totalPrice}</Text>
      <Text style={[styles.text, textStyle]}>Status: {status}</Text>
      <TouchableOpacity
        style={styles.updateButton}
        onPress={() => onUpdateOrder(orderId)}>
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
