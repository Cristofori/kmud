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

print '%s: Connecting...' % user

telnet = pexpect.spawn(telnetCommand, timeout=5)

def login(user, password):
    patterns = telnet.compile_pattern_list(['> $', 'Username: $', 'Password: $', 'already online', 'User not found', pexpect.TIMEOUT])

    while True:
        try:
            index = telnet.expect(patterns)
        except pexpect.EOF:
            print '%s: Lost connection to server' % user
            exit(0)

        if index == 0:
            telnet.sendline("l")
        elif index == 1:
            print '%s: Logging in' % user
            telnet.sendline(user)
        elif index == 2:
            print '%s: Sending password' % user
            telnet.sendline(password)
            telnet.sendline("1")
            break
        elif index == 3:
            print '%s: Login failed, already online' % user
            sys.exit(2)
        elif index == 4:
            print '%s: User not found' % user
            sys.exit(3)
        else:
            print '%s: Login timeout' % user
            break

def runaround():
    print '%s: Running around' % user
    exits = re.compile('Exits: (\[N\]orth)? ?(\[NE\]North East)? ?(\[E\]ast)? ?(\[SE\]South East)? ?(\[S\]outh)? ?(\[SW\]South West)? ?(\[W\]est)? ?(\[NW\]North West)?')
    patterns = telnet.compile_pattern_list([exits, pexpect.TIMEOUT])

    while True:
        try:
            index = telnet.expect(patterns)
        except pexpect.EOF:
            print '%s: Lost connection to server' % user
            exit(0)

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
            # print "%s: Moving %s" % (user, direction)
            telnet.sendline(direction)

        elif index == 1:
            print '%s: Runaround timeout' % user
            print telnet.before
            print telnet.after
            telnet.sendline("l")


login(user, password)
runaround()

# vim: nocindent

