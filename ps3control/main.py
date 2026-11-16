#!/usr/bin/env python3
"""
PS3 Controller Input Reader
Reads and responds to PS3 controller inputs using evdev on Linux
"""

import evdev
from evdev import InputDevice, categorize, ecodes
import sys

def find_ps3_controller():
    """Find the PS3 controller device"""
    devices = [InputDevice(path) for path in evdev.list_devices()]
    
    for device in devices:
        if "PLAYSTATION(R)3" in device.name or "Sony" in device.name:
            return device
    
    # If not found by name, list all devices for user to choose
    print("PS3 Controller not found automatically. Available devices:")
    for i, device in enumerate(devices):
        print(f"{i}: {device.path} - {device.name}")
    
    if not devices:
        print("No input devices found!")
        return None
    
    try:
        choice = int(input("Select device number: "))
        return devices[choice]
    except (ValueError, IndexError):
        print("Invalid selection")
        return None

def main():
    # Find and connect to controller
    controller = find_ps3_controller()
    
    if not controller:
        print("Could not find controller")
        sys.exit(1)
    
    print(f"Connected to: {controller.name}")
    print(f"Device path: {controller.path}")
    print("\nListening for inputs (Ctrl+C to exit)...\n")
    
    # Dictionary to track analog stick and trigger values
    state = {}
    
    try:
        for event in controller.read_loop():
            # Only process key and absolute axis events
            if event.type == ecodes.EV_KEY:
                # Button presses
                key_event = categorize(event)
                button_name = ecodes.BTN[event.code] if event.code in ecodes.BTN else f"BTN_{event.code}"
                
                if event.value == 1:  # Button pressed
                    print(f"Button pressed: {button_name}")
                    handle_button_press(button_name)
                elif event.value == 0:  # Button released
                    print(f"Button released: {button_name}")
            
            elif event.type == ecodes.EV_ABS:
                # Analog sticks and triggers
                abs_event = categorize(event)
                axis_name = ecodes.ABS[event.code] if event.code in ecodes.ABS else f"ABS_{event.code}"
                
                # Store current value
                state[axis_name] = event.value
                
                # Print significant changes (filter noise)
                if abs(event.value) > 10 or event.value == 0:
                    print(f"Axis {axis_name}: {event.value}")
                    handle_axis_change(axis_name, event.value, state)
    
    except KeyboardInterrupt:
        print("\n\nExiting...")
    finally:
        controller.close()

def handle_button_press(button_name):
    """Handle button press events - customize this for your needs"""
    
    # Example mappings for PS3 controller
    actions = {
        'BTN_SOUTH': 'Cross/X button - Action',
        'BTN_EAST': 'Circle button - Back',
        'BTN_NORTH': 'Triangle button - Special',
        'BTN_WEST': 'Square button - Attack',
        'BTN_TL': 'L1 - Left bumper',
        'BTN_TR': 'R1 - Right bumper',
        'BTN_TL2': 'L2 - Left trigger',
        'BTN_TR2': 'R2 - Right trigger',
        'BTN_SELECT': 'Select button',
        'BTN_START': 'Start button',
        'BTN_MODE': 'PS button',
        'BTN_THUMBL': 'Left stick click',
        'BTN_THUMBR': 'Right stick click',
    }
    
    if button_name in actions:
        print(f"  -> {actions[button_name]}")
    
    # Add your custom logic here
    # if button_name == 'BTN_SOUTH':
    #     # Do something when X is pressed
    #     pass

def handle_axis_change(axis_name, value, state):
    """Handle analog stick and trigger changes - customize this for your needs"""
    
    # Example: detect stick directions
    if axis_name == 'ABS_X':  # Left stick horizontal
        if value < -100:
            print("  -> Left stick: LEFT")
        elif value > 100:
            print("  -> Left stick: RIGHT")
    
    elif axis_name == 'ABS_Y':  # Left stick vertical
        if value < -100:
            print("  -> Left stick: UP")
        elif value > 100:
            print("  -> Left stick: DOWN")
    
    elif axis_name == 'ABS_RX':  # Right stick horizontal
        if value < -100:
            print("  -> Right stick: LEFT")
        elif value > 100:
            print("  -> Right stick: RIGHT")
    
    elif axis_name == 'ABS_RY':  # Right stick vertical
        if value < -100:
            print("  -> Right stick: UP")
        elif value > 100:
            print("  -> Right stick: DOWN")
    
    # Add your custom logic here
    # Example: use left stick for movement
    # left_x = state.get('ABS_X', 0)
    # left_y = state.get('ABS_Y', 0)
    # move_character(left_x, left_y)

if __name__ == "__main__":
    main()
