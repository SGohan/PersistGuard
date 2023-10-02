# PersistGuard
PersistGuard is a golang program responsible for providing us persistence in Windows through RDP

**Usage**

```go build main.go```

You have to open a CMD as administration privileges.

```main.exe```

Now you'll have to enter the path where you have the main.exe located (to create the scheduled task), also you'll have to enter administrators group name and remote desktop users group name.

![image](https://github.com/a11cyberbull/PersistGuard/assets/103254517/2d63c39d-b400-497e-80c2-36bccd939c68)

Finally youll have RDP activated, an scheduled task created that will execute the program each hour, and "support" user created with password "P@ssw0rd!" so you can access remotely with that user.

**Revert process**

To revert the process and delete the scheduled task, delete support user, and close the firewall rule, we can execute the revert.go and make the process.

```go build revert.go```

You have to open a CMD as administration privileges.

```revert.exe```

Now you'll have to enter the path where you have the main.exe located (to kill the scheduled task),

![image](https://github.com/a11cyberbull/PersistGuard/assets/103254517/c4f90633-d636-4128-a210-3f51c3234e37)

Done! 



