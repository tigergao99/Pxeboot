# Pxeboot
The script does the following: 
1. Build from git url.
2. Hang idle if test machine is up. Proceed if it's down. This is to handle concurrency.
3. Clean up old system.
4. Copy over new system and configuration.
5. Boot up test machine.
6. Connect to test machine and run kyua test.
The git repo and jailname can be specified using the test.json file.
### To run
You can either: 
* run `./testbuild` directly
Or:
* go run testbuild.go

