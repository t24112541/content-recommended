### Single User Load Test

```
k6 run --env K6_SCENARIO_NAME=single_user_load test/test.js
```

### Batch Stress Test

```
k6 run --env K6_SCENARIO_NAME=batch_stress test/test.js
```

### Cache Effectiveness Test

```
k6 run --env K6_SCENARIO_NAME=cache_effectiveness test/test.js
```

### All Scenarios

```
k6 run test/test.js
```
