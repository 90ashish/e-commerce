import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  vus: 50,           // 50 virtual users
  duration: '30s',   // for 30 seconds
};

export default function () {
  const orderID = `o-${__VU}-${Date.now()}`;
  const payload = JSON.stringify({
    order_id: orderID,
    user_id: `u-${__VU}`,
    items: ['foo'],
    total: 9.99,
  });
  const params = { headers: { 'Content-Type': 'application/json' } };

  let res = http.post('http://localhost:8090/orders', payload, params);
  check(res, { 'status is 202': (r) => r.status === 202 });
  sleep(0.1);
}
