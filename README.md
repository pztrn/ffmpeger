# FFMpeger

Simple microservice to convert video files with ffmpeg and receiving commands via NATS.

## How to use

1. Launch docker-compose to start required services.
2. Start ``ffmpeger.go`` from ``cmd/ffmpeger``. Please take a look at help (``-h``)!
3. Launch example message sender from ``cmd/send_example_message`` specifying input and output video files paths. See help (``-h``).