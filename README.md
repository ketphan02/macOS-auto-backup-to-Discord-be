# Backend

## Current capability
- Able to split a file into multiple chunks
- Encryption/Decryption data with passkey
- Send the encrypted data to the discord server
- Receive the encrypted data from the discord server and decrypt, merge, and save it to local directory

## Future capability
- Support deletion of the file
- Modify file name if duplicated
- Local files should be arranged in the same hierarchy as the original files
- Allow to set time-of-life for data


## Run Development
Create a `.env` file in the root directory with content as in `.env.example` file.

To initialize the database, run the following command:
```shell
docker-compose up db -d
```

To start the development server, run the following command:
```shell
task :start -w --interval=500ms
```
