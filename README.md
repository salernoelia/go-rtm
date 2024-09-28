ffmpeg -f avfoundation -pixel_format nv12 -framerate 30 -video_size 1280x720 -i "0" -c:v libx264 -preset ultrafast -tune zerolatency -f rtsp -rtsp_transport tcp rtsp://localhost:8554/mystream

```yml
# mediamtx.yml
paths:
  # example:
  # my_camera:
  #   source: rtsp://my_camera

  my_stream:
    source: rtsp://my_stream
```

url:

```
rtsp://localhost:8554/mystream
```

Launch Server

```
rtsp-server/mediamtx rtsp-server/mediamtx.yml
```
