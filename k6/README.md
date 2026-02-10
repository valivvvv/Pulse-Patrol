# k6 Performance Tests

Start the server first:

```bash
go run .
```

Then run one of the three test types:

```bash
k6 run --env TEST_TYPE=load k6/test.js
k6 run --env TEST_TYPE=stress k6/test.js
k6 run --env TEST_TYPE=spike k6/test.js
```
