# E2E component test

1. Create python virtual env
   ```
   ./create_env.sh
   ```
2. Build test image (docker)
   ```
   ./build.sh
   ```
3. Execute test
   ```
   test_env/bin/python component_test.py --image gateway:test
   ```