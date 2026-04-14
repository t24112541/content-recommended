```
k6 run test/test.js

         /\      Grafana   /‾‾/
    /\  /  \     |\  __   /  /
   /  \/    \    | |/ /  /   ‾‾\
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/


     execution: local
        script: test/test.js
        output: -

     scenarios: (100.00%) 3 scenarios, 320 max VUs, 2m30s max duration (incl. graceful stop):
              * batch_stress: Up to 100 looping VUs for 2m0s over 3 stages (gracefulRampDown: 30s, gracefulStop: 30s)
              * cache_effectiveness: 20 looping VUs for 1m0s (gracefulStop: 30s)
              * single_user_load: 100.00 iterations/s for 1m0s (maxVUs: 50-200, gracefulStop: 30s)



  █ THRESHOLDS

    http_req_duration
    ✓ 'p(95)<500' p(95)=186.63ms
    ✓ 'p(99)<1000' p(99)=530.23ms

      {scenario:batch_stress}
      ✓ 'p(95)<800' p(95)=164.51ms

      {scenario:cache_effectiveness}
      ✓ 'p(95)<300' p(95)=209.76ms

      {scenario:single_user_load}
      ✓ 'p(95)<500' p(95)=477.99ms

    http_req_failed
    ✓ 'rate<0.01' rate=0.00%


  █ TOTAL RESULTS

    checks_total.......: 120632 1004.812185/s
    checks_succeeded...: 99.99% 120626 out of 120632
    checks_failed......: 0.00%  6 out of 120632

    ✗ status is 200
      ↳  99% — ✓ 60313 / ✗ 3
    ✗ has recommendations
      ↳  99% — ✓ 60313 / ✗ 3

    HTTP
    http_req_duration....................: avg=47.88ms  min=506µs    med=13.36ms  max=3.37s p(90)=127.59ms p(95)=186.63ms
      { expected_response:true }.........: avg=47.86ms  min=506µs    med=13.36ms  max=3.37s p(90)=127.51ms p(95)=186.34ms
      { scenario:batch_stress }..........: avg=41.43ms  min=515µs    med=13ms     max=3.37s p(90)=116.55ms p(95)=164.51ms
      { scenario:cache_effectiveness }...: avg=51.88ms  min=519.8µs  med=14.98ms  max=1.37s p(90)=138.18ms p(95)=209.76ms
      { scenario:single_user_load }......: avg=93.27ms  min=506µs    med=16.05ms  max=1.29s p(90)=275.85ms p(95)=477.99ms
    http_req_failed......................: 0.00%  3 out of 60316
    http_reqs............................: 60316  502.406093/s

    EXECUTION
    dropped_iterations...................: 58     0.483115/s
    iteration_duration...................: avg=150.78ms min=100.65ms med=115.94ms max=3.49s p(90)=230.6ms  p(95)=291.05ms
    iterations...........................: 60316  502.406093/s
    vus..................................: 1      min=1          max=134
    vus_max..............................: 208    min=170        max=208

    NETWORK
    data_received........................: 93 MB  777 kB/s
    data_sent............................: 6.2 MB 52 kB/s




running (2m00.1s), 000/208 VUs, 60316 complete and 0 interrupted iterations
batch_stress        ✓ [======================================] 000/100 VUs  2m0s
cache_effectiveness ✓ [======================================] 20 VUs       1m0s
single_user_load    ✓ [======================================] 000/088 VUs  1m0s  100.00 iters/s

```
