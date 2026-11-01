import http from 'k6/http';
import { check } from 'k6';
import { Counter } from 'k6/metrics';
export const options = {
  scenarios: { burst: { executor: 'constant-arrival-rate', rate: 500, timeUnit: '1s', duration: '20s', preAllocatedVUs: 100, maxVUs: 1000 } },
  thresholds: { http_req_failed: ['rate<0.01'], http_req_duration: ['p(95)<50'] }
};
const allowed = new Counter('allowed');
const denied = new Counter('denied');
export default function () {
  const r = http.post(`${__ENV.BASE_URL || 'http://localhost:8080'}/v1/rate-limit/check`,
    JSON.stringify({client_id: __ENV.CLIENT_ID || 'demo-client'}),
    {headers:{'Content-Type':'application/json'}});
  if (r.status === 200) allowed.add(1);
  if (r.status === 429) denied.add(1);
  check(r, {'status is ALLOW or DENY': x => x.status === 200 || x.status === 429});
}
