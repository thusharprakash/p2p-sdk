import {computeOrderState} from '@oolio-group/order-helper';

export const globalCache = {};

export function addGlobalEvents(orderId, events) {
  var newEvents = [];
  const existingEvents = globalCache[orderId];
  if (existingEvents === undefined || existingEvents.length === 0) {
    newEvents.push(...events);
  } else {
    newEvents.push(...existingEvents);
    for (let i = 0; i < events.length; i++) {
      const event = events[i];
      const isExists = existingEvents.filter(e => e.id === event.id).length > 0;
      if (!isExists) {
        newEvents.push(event);
      }
    }
  }
  newEvents = newEvents.sort((a, b) => a.timestamp - b.timestamp);
  globalCache[orderId] = newEvents;
}

export function generateOrdersFromCache() {
  const orders = [];
  for (const key in globalCache) {
    const events = globalCache[key];
    const order = computeOrderState(events);
    orders.push({
      id: order.id,
      total: order.totalPrice,
    });
  }
  return orders;
}

export function getLastEvent(orderId) {
  const events = globalCache[orderId];
  if (events === undefined || events.length === 0) {
    return null;
  }
  return events[events.length - 1].id;
}

export function generateFullOrdersFromCache() {
  const orders = [];
  for (const key in globalCache) {
    const events = globalCache[key];
    const order = computeOrderState(events);
    orders.push(order);
  }
  return orders;
}
