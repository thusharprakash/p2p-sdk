import {computeOrderState} from '@oolio-group/order-helper';

export const globalCache = {};

export function addGlobalEvents(orderId, events) {
  try {
    var newEvents = [];
    const existingEvents = globalCache[orderId]?.events || undefined;
    let changed = false;
    if (existingEvents === undefined || existingEvents.length === 0) {
      newEvents.push(...events);
      changed = true;
    } else {
      newEvents.push(...existingEvents);
      for (let i = 0; i < events.length; i++) {
        const event = events[i];
        const isExists =
          existingEvents.filter(e => e.id === event.id).length > 0;
        if (!isExists) {
          console.log('Adding new event to cache');
          newEvents.push(event);
          changed = true;
        }
      }
    }
    if (changed === true) {
      newEvents = newEvents.sort((a, b) => a.timestamp - b.timestamp);
      globalCache[orderId] = {
        events: newEvents,
        order: undefined,
      };
    }
  } catch (e) {
    console.log(e);
  }
}

export function generateOrdersFromCache() {
  const orders = [];
  for (const key in globalCache) {
    const events = globalCache[key].events;
    const order = computeOrderState(events);
    orders.push({
      id: order.id,
      total: order.totalPrice,
    });
  }
  return orders;
}

export function getLastEvent(orderId) {
  const events = globalCache[orderId].events;
  if (events === undefined || events.length === 0) {
    return null;
  }
  return events[events.length - 1].id;
}

export function generateFullOrdersFromCache(previousOrders) {
  const orders = [];
  for (const key in globalCache) {
    const events = globalCache[key].events;
    if (globalCache[key].order === undefined) {
      const order = computeOrderState(events, previousOrders[key]);
      orders.push(order);
      globalCache[key].order = order;
    } else {
      orders.push(globalCache[key].order);
    }
  }
  return orders;
}

export function areObjectsEqual(obj1, obj2) {
  // Check if all keys and their corresponding values are equal
  for (const item of obj1) {
    try {
      const key = item.id;

      const isIDEqual = item.id === obj2[key].id;
      const isPriceEqual = item.totalPrice === obj2[key].totalPrice;
      const isStatusEqual = item.status === obj2[key].status;
      if (isIDEqual && isPriceEqual && isStatusEqual) {
        continue;
      } else {
        return false;
      }
    } catch (e) {
      console.log(e);
      return false;
    }
  }

  return true;
}
