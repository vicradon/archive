# PS3 control

This repo has a Python and C program for running PS control checks. To use, ensure the following is installed:

1. build-essential
2. evdev
3. linux-headers-$(uname -r)
4. joystick
5. joystick-gtk

## Messing with bluetoothctl 

Generally, when you run the ctl command, you run these:

```sh
agent on
default-agent
scan on
connect <mac address>
```


