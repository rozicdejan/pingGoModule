# 1 create .env file with

TELEGRAM_BOT_TOKEN=your_bot_token
TELEGRAM_CHAT_ID=your_chat_id

# 2 Edit device.yaml file
    add description
    add IP addr
# 3 run ping_monitor or. ping_monitor.exe



# Running on Linux
## Prepare your environment:

 Make sure your .env and devices.yaml files are in the same directory as the ping_monitor executable.
Make the executable runnable (if necessary):

    chmod +x ping_monitor
## Run the application:
    ./ping_monitor
## Running in the Background on Linux
    nohup ./ping_monitor &
or
    ./ping_monitor &
## Managing the Application
    ps aux | grep ping_monitor
    kill <PID>
## Logging and Output
Linux: You can redirect the output to a log file for later review:
    
    nohup ./ping_monitor > ping_monitor.log 2>&1 &
    
