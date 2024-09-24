### Replayor

This is a very rough, very WIP tool for replaying blocks on an op-stack network and outputting engine API timing information.

It uses a fork of [op-geth](https://github.com/ethereum-optimism/op-geth/compare/optimism...danyalprout:op-geth:danyal-wip?expand=1) which contains a hack to drop individual failed transactions from the block, instead of the whole block. This is necessary for replaying blocks with different parameters as some of the original transactions may fail due to the change in parameters.

### Run a test

```bash
# Initialize the engine jwt
make init

# Copy a snapshot which you want to test on top of
# note you can run this without a snapshot from genesis, but it's less effective for testing
cp /path/to/snapshot /path/to/replayor/geth-data-archive

# Configure your test parameters in an env file in 
vim test-configs/my-test.env

# Update the components to use this test information
# by changing the env_file properties in the docker-compose.yml
vim docker-compose.yml

# Run the test
make run

# If using local file system for results, you can view them in the results/ directory
```