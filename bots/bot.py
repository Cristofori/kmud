#!/usr/bin/python

import pexpect, sys

#host = sys.argv[1]
#port = sys.argv[2]
#user = sys.argv[3]
#password = sys.argv[4]
host = "localhost"
port = 8945
user = "chris"
password = "asdf"

telnetCommand = 'telnet %s %s' % (host, port)

print 'Connecting...'
telnet = pexpect.spawn(telnetCommand, timeout=5)

def login(user, password):
    print 'Logging in as %s' % user
    patterns = telnet.compile_pattern_list(['> $', 'Username: $', 'Password: $', 'already online', pexpect.TIMEOUT])

    while True:
        index = telnet.expect(patterns)

        if index == 0:
            print 'Sending L'
            telnet.sendline("l")
        elif index == 1:
            print 'Sending username: %s' % user
            telnet.sendline(user)
        elif index == 2:
            print 'Sending password: %s' % password
            telnet.sendline(password)
            telnet.sendline("1")
            break
        elif index == 3:
            print 'Login failed, %s is already online' % user
            break
        else:
            print 'Timeout'
            break


login(user, password)

# vim: nocindent

