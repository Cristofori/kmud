#!/usr/bin/python

import pexpect, sys, re, random

#host = sys.argv[1]
#port = sys.argv[2]
#user = sys.argv[3]
#password = sys.argv[4]

host = "localhost"
port = 8945
user = "chris"
password = "a"

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
            sys.exit(2)
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

