import uuid from 'react-native-uuid';
import {
  order1Events,
  order2Events,
  order3Events,
  updateWithItems,
} from './sample';

export const generateOrder = id => {
  // Generate a random order id as uuid
  const orderId = uuid.v4();

  // Get an order at random from the list of order examples
  const orderExamples = [order1Events, order2Events, order3Events];
  const order = orderExamples[id];
  const updatedEvents = order.map(event => {
    return {
      ...event,
      orderId: orderId,
    };
  });

  return updatedEvents;
};

export const updateOrder = (id, previous) => {
  console.log('update order', id);
  const itemEvent = updateWithItems();
  return itemEvent.map(event => {
    return {
      ...event,
      orderId: id,
      previous: previous,
    };
  });
};
