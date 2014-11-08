**Problem:** You want to accept connections on a port number under 1024 on OS X, but you don't want your server to run as root.

**Solution:** Run your server with _launch_socket_server_.

_launch_socket_server_ sits in between [launchd(8)](https://developer.apple.com/library/mac/documentation/Darwin/Reference/ManPages/man8/launchd.8.html) and your server, using the system's APIs to proxy incoming connections from a privileged port to your unprivileged program.

<img src="https://raw.githubusercontent.com/sstephenson/launch_socket_server/master/share/launch_socket_server/launch_socket_server.png?token=AAAKK38n5g033WRkEEbwF3J5ADQhQ9rjks5UalzqwA%3D%3D">

It uses the `launch_activate_socket` API, available on OS X 10.9 and higher.

_launch_socket_server_ can proxy incoming connections to a [Unix domain socket](http://en.wikipedia.org/wiki/Unix_domain_socket) or to a local TCP port. By default, _launch_socket_server_ generates the path to a domain socket and passes that path to your program in an environment variable.

To run your server with _launch_socket_server_, begin by creating a launch daemon.

---

#### Creating and configuring a conventional launch daemon

The following configuration sets up launchd to listen on port 80. _launch_socket_server_ activates the port, runs your server program, and proxies requests to it via a UNIX domain socket. Your server is responsible for reading the path to the domain socket from the `LAUNCH_PROGRAM_SOCKET_PATH` environment variable and accepting connections on that socket.

1. Pick an identifier for your server. By convention, launch daemon identifiers start with the components of a domain name you control, in reverse order, followed by the name of your server. For this example, we will use an identifier of `com.example.myserver`.

2. Create an XML plist file at `/Library/LaunchDaemons/<identifier>.plist`, replacing _identifier_ in the pathname with your server's identifier.

3. In the plist file, specify your server's identifier, tell it to execute when loaded, and request that launchd keeps it running.
   ```xml
   <?xml version="1.0" encoding="UTF-8"?>
   <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
   <plist version="1.0">
   <dict>
       <key>Label</key>
       <string>com.example.myserver</string>
       <key>RunAtLoad</key>
       <true/>
       <key>KeepAlive</key>
       <true/>
   ```

4. Specify the user and group you want to use to run your server program.
   ```xml
       <key>UserName</key>
       <string>sam</string>
       <key>GroupName</key>
       <string>staff</string>
   ```

5. Tell launchd where to redirect your server program's stdout and stderr. (Optional, but recommended.)
   ```xml
       <key>StandardOutPath</key>
       <string>/usr/local/var/log/myserver.log</string>
       <key>StandardErrorPath</key>
       <string>/usr/local/var/log/myserver.log</string>
   ```

6. Specify the path to the `launch_socket_server` executable, your server program, and any arguments you wish to pass to your server.
   ```xml
       <key>ProgramArguments</key>
       <array>
           <!-- path to launch_socket_server -->
           <string>/usr/local/sbin/launch_socket_server</string>
           <!-- path to your server program, and any arguments you wish to pass -->
           <string>/usr/local/bin/myserver</string>
           <string>argument1</string>
           <string>argument2</string>
       </array>
   ```

7. Define a launch daemon socket with the address and port you want to listen on. (Note: The _SockServiceName_ value may be a port number, or the name of any service defined in `/etc/services`.)
   ```xml
       <key>Sockets</key>
       <dict>
           <key>Socket</key>
           <dict>
               <key>SockNodeName</key>
               <string>0.0.0.0</string>
               <key>SockServiceName</key>
               <string>80</string>
           </dict>
       </dict>
   </dict>
   </plist>
   ```

#### Configuring your server to run with launch_socket_server

Update your server's configuration to bind to the UNIX domain socket specified in the `LAUNCH_PROGRAM_SOCKET_PATH` environment variable, if it is set. Your server must unlink the socket before binding if the socket already exists, and should unlink the socket when the program terminates.

For more information on using domain sockets, please see the documentation for your programming environment.

Note: Your server program must run in the foreground, not forked as a background process.

##### Manually specifying the path to the domain socket

If you wish to specify the path to the domain socket shared by _launch_socket_server_ and your program, you may set the `LAUNCH_PROGRAM_SOCKET_PATH` environment variable in the launch daemon plist file.

```xml
    <key>EnvironmentVariables</key>
    <dict>
        <key>LAUNCH_PROGRAM_SOCKET_PATH</key>
        <string>/tmp/myserver.sock</string>
    </dict>
```

##### Proxying to a TCP address instead of a domain socket

If you wish to proxy requests to a local TCP address on an unprivileged port, you may set the `LAUNCH_PROGRAM_TCP_ADDRESS` environment variable in the launch daemon plist file. The value must be the destination address and port separated by a colon.

```xml
    <key>EnvironmentVariables</key>
    <dict>
        <key>LAUNCH_PROGRAM_TCP_ADDRESS</key>
        <string>127.0.0.1:8000</string>
    </dict>
```

##### Proxying without running a server program

You may wish to proxy requests to a domain socket or TCP address without having _launch_socket_server_ run your server program. To do this, first set either `LAUNCH_PROGRAM_SOCKET_PATH` or `LAUNCH_PROGRAM_TCP_ADDRESS` as described above, and then specify a `-` in place of the server program path.

```xml
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/sbin/launch_socket_server</string>
        <string>-</string>
    </array>
```

##### Running your server program with the user's login environment

Your server program will run with a minimal environment. If your program relies on certain environment variables (such as `PATH`) being set by the user's shell profile, you may first pass your program through the [`login_wrapper`](libexec/launch_socket_server/login_wrapper) helper, which re-executes your program through the user's login shell.

```xml
    <key>ProgramArguments</key>
    <array>
        <!-- path to launch_socket_server -->
        <string>/usr/local/sbin/launch_socket_server</string>
        <!-- path to login_wrapper -->
        <string>/usr/local/libexec/launch_socket_server/login_wrapper</string>
        <!-- path to your server program, and any arguments you wish to pass -->
        <string>/usr/local/bin/myserver</string>
        <string>argument1</string>
        <string>argument2</string>
    </array>
```


#### Registering your launch daemon with launchd

To register and load your launch daemon, use the `launchctl load` command as root:

```
$ sudo launchctl load -Fw /Library/LaunchDaemons/com.example.myserver.plist
```

To unload your launch daemon, use the `launchctl unload` command as root:

```
$ sudo launchctl unload -Fw /Library/LaunchDaemons/com.example.myserver.plist
```

---

#### Building and installing launch_socket_server

To build _launch_socket_server_ from source, first install Go, then run the `make` command inside a copy of the source tree. (You can install Go using [Homebrew](http://brew.sh/).)

```
$ brew install go
$ make
```

To install into `/usr/local`, run `make install`.

```
$ make install
```

Set the `PREFIX` environment variable to install into a different location.

---

Copyright © 2014 Sam Stephenson <<sstephenson@gmail.com>>

Freely distributable under the [MIT license](LICENSE)
