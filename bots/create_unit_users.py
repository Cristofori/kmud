#!/usr/bin/python

import pexpect, sys, exceptions

host = "localhost"
port = 8945

baseUsername = "unit"
password = "unit123"

def usage():
    print "Usage: %s [count] [host] [port]" % sys.argv[0]
    sys.exit(1)

if len(sys.argv) < 2:
    usage()

try:
    userCount = int(sys.argv[1])
except exceptions.ValueError:
    usage()

if len(sys.argv) > 2:
    host = sys.argv[2]

if len(sys.argv) > 3:
    try:
        port = int(sys.argv[3])
    except exceptions.ValueError:
        usage()

telnetCommand = 'telnet %s %s' % (host, port)

print 'Connecting...'
telnet = pexpect.spawn(telnetCommand, timeout=5)

patterns = telnet.compile_pattern_list(['> $', 'Desired username', 'unavailable', 'Desired password', 'Confirm password', pexpect.TIMEOUT])

for i in range(userCount):
    username = baseUsername + str(i+1)
    while True:
        index = telnet.expect(patterns)

        if index == 0:
            telnet.sendline("n")
        elif index == 1:
            telnet.sendline(username)
        elif index == 2:
            break # User already exists, move on
        elif index == 3:
            telnet.sendline(password)
        elif index == 4:
            telnet.sendline(password)

            while True:
                charPatterns = telnet.compile_pattern_list(['> $', 'Desired character name', pexpect.TIMEOUT])
                index = telnet.expect(charPatterns)

                if index == 0:
                    telnet.sendline("n")
                elif index == 1:
                    telnet.sendline(username)
                    telnet.sendline("x")
                    telnet.sendline("x")
                    break

            break

        else:
            print 'Timeout, %s' % username
            break

# vim: nocindent

