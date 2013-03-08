#!/usr/bin/python

import pexpect, sys, re, random, exceptions

host = "localhost"
port = 8945
user = "unit"
password = "unit"

def usage():
    print "Usage: %s [host] [port] [username] [password]" % sys.argv[0]
    sys.exit(1)

if len(sys.argv) > 1:
    host = sys.argv[1]

if len(sys.argv) > 2:
    try:
        port = int(sys.argv[2])
    except exceptions.ValueError:
        usage()

if len(sys.argv) > 3:
    user = sys.argv[3]

if len(sys.argv) > 4:
    password = sys.argv[4]

telnetCommand = 'telnet %s %s' % (host, port)

print 'Connecting...'
telnet = pexpect.spawn(telnetCommand, timeout=5)

def login(user, password):
    print 'Logging in as %s' % user
    patterns = telnet.compile_pattern_list(['> $', 'Username: $', 'Password: $', 'already online', 'User not found', pexpect.TIMEOUT])

    while True:
        index = telnet.expect(patterns)

        if index == 0:
            telnet.sendline("l")
        elif index == 1:
            print 'Logging in as %s' % user
            telnet.sendline(user)
        elif index == 2:
            print 'Sending password: %s' % password
            telnet.sendline(password)
            telnet.sendline("1")
            break
        elif index == 3:
            print 'Login failed, %s is already online' % user
            sys.exit(2)
        elif index == 4:
            print 'User not found: %s' % user
            sys.exit(3)
        else:
            print 'Login timeout'
            break

def runaround():
    print 'Running around'
    exits = re.compile('Exits: (\[N\]orth)? ?(\[NE\]North East)? ?(\[E\]ast)? ?(\[SE\]South East)? ?(\[S\]outh)? ?(\[SW\]South West)? ?(\[W\]est)? ?(\[NW\]North West)?')
    patterns = telnet.compile_pattern_list([exits, pexpect.TIMEOUT])

    while True:
        index = telnet.expect(patterns)

        if index == 0:
            m = telnet.match

            exitList = []

            for i in range(len(m.groups())):
                cap = m.group(i)
                if cap != None:
                    if i == 1: # N
                        exitList.append("N")
                    elif i == 2: # NE
                        exitList.append("NE")
                    elif i == 3: # E
                        exitList.append("E")
                    elif i == 4: # SE
                        exitList.append("SE")
                    elif i == 5: # S
                        exitList.append("S")
                    elif i == 6: # SW
                        exitList.append("SW")
                    elif i == 7: # W
                        exitList.append("W")
                    elif i == 8: # NW
                        exitList.append("NW")

            index = random.randint(0, len(exitList) - 1)
            direction = exitList[index]
            print "Moving %s" % direction
            telnet.sendline(direction)

        elif index == 1:
            print 'Runaround timeout'
            telnet.sendline("l")


login(user, password)
runaround()

# vim: nocindent

