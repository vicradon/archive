#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h>
#include <linux/input.h>
#include <dirent.h>

#define INPUT_DIR "/dev/input"
#define MAX_PATH 256

// Function prototypes
int find_ps3_controller(char *device_path);
void handle_button_event(struct input_event *ev);
void handle_axis_event(struct input_event *ev);
const char* get_button_name(int code);
const char* get_axis_name(int code);

int main() {
    char device_path[MAX_PATH];
    int fd;
    struct input_event ev;
    
    // Find PS3 controller
    if (find_ps3_controller(device_path) < 0) {
        fprintf(stderr, "Could not find PS3 controller\n");
        return 1;
    }
    
    printf("Found controller: %s\n", device_path);
    
    // Open device
    fd = open(device_path, O_RDONLY);
    if (fd < 0) {
        perror("Failed to open device");
        return 1;
    }
    
    printf("Listening for inputs (Ctrl+C to exit)...\n\n");
    
    // Read events loop
    while (read(fd, &ev, sizeof(struct input_event)) > 0) {
        switch (ev.type) {
            case EV_KEY:
                handle_button_event(&ev);
                break;
            case EV_ABS:
                handle_axis_event(&ev);
                break;
        }
    }
    
    close(fd);
    return 0;
}

int find_ps3_controller(char *device_path) {
    DIR *dir;
    struct dirent *entry;
    char path[MAX_PATH];
    char name[256];
    int fd;
    
    dir = opendir(INPUT_DIR);
    if (!dir) {
        perror("Could not open /dev/input");
        return -1;
    }
    
    // Look for event devices
    while ((entry = readdir(dir)) != NULL) {
        if (strncmp(entry->d_name, "event", 5) != 0)
            continue;
        
        snprintf(path, MAX_PATH, "%s/%s", INPUT_DIR, entry->d_name);
        fd = open(path, O_RDONLY);
        if (fd < 0)
            continue;
        
        // Get device name
        if (ioctl(fd, EVIOCGNAME(sizeof(name)), name) >= 0) {
            // Check if it's a PS3 controller
            if (strstr(name, "PLAYSTATION(R)3") || 
                strstr(name, "Sony") ||
                strstr(name, "PS3")) {
                printf("Found: %s - %s\n", path, name);
                strcpy(device_path, path);
                close(fd);
                closedir(dir);
                return 0;
            }
        }
        close(fd);
    }
    
    closedir(dir);
    
    // If not found, list available devices
    printf("PS3 controller not found. Available devices:\n");
    dir = opendir(INPUT_DIR);
    int i = 0;
    while ((entry = readdir(dir)) != NULL) {
        if (strncmp(entry->d_name, "event", 5) != 0)
            continue;
        
        snprintf(path, MAX_PATH, "%s/%s", INPUT_DIR, entry->d_name);
        fd = open(path, O_RDONLY);
        if (fd >= 0) {
            if (ioctl(fd, EVIOCGNAME(sizeof(name)), name) >= 0) {
                printf("%d: %s - %s\n", i++, path, name);
            }
            close(fd);
        }
    }
    closedir(dir);
    
    return -1;
}

void handle_button_event(struct input_event *ev) {
    if (ev->value == 1) {
        printf("Button pressed: %s (code: %d)\n", get_button_name(ev->code), ev->code);
        
        // Add your custom logic here
        switch (ev->code) {
            case BTN_SOUTH:  // X/Cross
                printf("  -> Cross/X button - Action\n");
                break;
            case BTN_EAST:   // Circle
                printf("  -> Circle button - Back\n");
                break;
            case BTN_NORTH:  // Triangle
                printf("  -> Triangle button - Special\n");
                break;
            case BTN_WEST:   // Square
                printf("  -> Square button - Attack\n");
                break;
            case BTN_TL:     // L1
                printf("  -> L1 - Left bumper\n");
                break;
            case BTN_TR:     // R1
                printf("  -> R1 - Right bumper\n");
                break;
            case BTN_SELECT:
                printf("  -> Select button\n");
                break;
            case BTN_START:
                printf("  -> Start button\n");
                break;
            case BTN_MODE:
                printf("  -> PS button\n");
                break;
        }
    } else if (ev->value == 0) {
        printf("Button released: %s (code: %d)\n", get_button_name(ev->code), ev->code);
    }
}

void handle_axis_event(struct input_event *ev) {
    // Filter out noise (small movements near center)
    if (abs(ev->value) < 10 && ev->value != 0)
        return;
    
    printf("Axis %s: %d\n", get_axis_name(ev->code), ev->value);
    
    // Add your custom logic here
    switch (ev->code) {
        case ABS_X:  // Left stick horizontal
            if (ev->value < -100)
                printf("  -> Left stick: LEFT\n");
            else if (ev->value > 100)
                printf("  -> Left stick: RIGHT\n");
            break;
        case ABS_Y:  // Left stick vertical
            if (ev->value < -100)
                printf("  -> Left stick: UP\n");
            else if (ev->value > 100)
                printf("  -> Left stick: DOWN\n");
            break;
        case ABS_RX:  // Right stick horizontal
            if (ev->value < -100)
                printf("  -> Right stick: LEFT\n");
            else if (ev->value > 100)
                printf("  -> Right stick: RIGHT\n");
            break;
        case ABS_RY:  // Right stick vertical
            if (ev->value < -100)
                printf("  -> Right stick: UP\n");
            else if (ev->value > 100)
                printf("  -> Right stick: DOWN\n");
            break;
    }
}

const char* get_button_name(int code) {
    switch (code) {
        case BTN_SOUTH: return "BTN_SOUTH";
        case BTN_EAST: return "BTN_EAST";
        case BTN_NORTH: return "BTN_NORTH";
        case BTN_WEST: return "BTN_WEST";
        case BTN_TL: return "BTN_TL";
        case BTN_TR: return "BTN_TR";
        case BTN_TL2: return "BTN_TL2";
        case BTN_TR2: return "BTN_TR2";
        case BTN_SELECT: return "BTN_SELECT";
        case BTN_START: return "BTN_START";
        case BTN_MODE: return "BTN_MODE";
        case BTN_THUMBL: return "BTN_THUMBL";
        case BTN_THUMBR: return "BTN_THUMBR";
        default: return "UNKNOWN";
    }
}

const char* get_axis_name(int code) {
    switch (code) {
        case ABS_X: return "ABS_X";
        case ABS_Y: return "ABS_Y";
        case ABS_Z: return "ABS_Z";
        case ABS_RX: return "ABS_RX";
        case ABS_RY: return "ABS_RY";
        case ABS_RZ: return "ABS_RZ";
        default: return "UNKNOWN";
    }
}
