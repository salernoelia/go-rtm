package main

import (
	"log"
	"os"
	"os/exec"
	"time"

	"gocv.io/x/gocv"
)

func main() {
    // RTSP server URL
    rtspURL := "rtsp://192.168.1.86:8554/mystream"

    // Initialize webcam
    webcam, err := gocv.OpenVideoCapture(0)
    if err != nil {
        log.Fatalf("Error opening webcam: %v", err)
    }
    defer webcam.Close()

    // Set frame size and frame rate
    webcam.Set(gocv.VideoCaptureFrameWidth, 1280)
    webcam.Set(gocv.VideoCaptureFrameHeight, 720)
    webcam.Set(gocv.VideoCaptureFPS, 30)

    // Verify frame size and FPS
    width := webcam.Get(gocv.VideoCaptureFrameWidth)
    height := webcam.Get(gocv.VideoCaptureFrameHeight)
    fps := webcam.Get(gocv.VideoCaptureFPS)
    log.Printf("Webcam initialized with frame size: %dx%d and FPS: %.2f", int(width), int(height), fps)

    // Start the FFmpeg process with optimized parameters
    ffmpegCmd := exec.Command("ffmpeg",
        "-f", "rawvideo",              // Raw video input format
        "-pixel_format", "bgr24",      // OpenCV outputs BGR format
        "-video_size", "1280x720",     // Match your webcam resolution
        "-framerate", "30",            // Match your webcam FPS
        "-i", "-",                      // Input from stdin
        "-c:v", "libx264",             // Encode with H264
        "-preset", "veryfast",          // Faster preset for better quality
        "-tune", "zerolatency",         // Keep zerolatency for streaming
        "-pix_fmt", "yuv420p",          // Convert to yuv420p
        "-b:v", "2M",                   // Set a constant bitrate (2 Mbps)
        "-bufsize", "2M",               // Set buffer size
        "-maxrate", "2M",               // Set maximum bitrate
        "-fflags", "nobuffer",          // Reduce FFmpeg internal buffering
        "-flags", "low_delay",          // Enable low delay
        "-strict", "-2",
        "-probesize", "32",
        "-analyzeduration", "0",
        "-threads", "2",                // Use multiple threads
        "-f", "rtsp",
        "-rtsp_transport", "tcp",
        rtspURL,                        // Send to the RTSP URL
    )

    // Redirect FFmpeg's stderr and stdout to Go's stderr and stdout for monitoring
    ffmpegCmd.Stdout = os.Stdout
    ffmpegCmd.Stderr = os.Stderr

    stdin, err := ffmpegCmd.StdinPipe()
    if err != nil {
        log.Fatalf("Failed to create FFmpeg stdin pipe: %v", err)
    }

    // Start the FFmpeg command
    if err := ffmpegCmd.Start(); err != nil {
        log.Fatalf("Failed to start FFmpeg: %v", err)
    }

    // Create a Mat to hold each frame
    img := gocv.NewMat()
    defer img.Close()

    // Use a ticker to maintain frame rate
    ticker := time.NewTicker(time.Millisecond * 33) // ~30 FPS
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if ok := webcam.Read(&img); !ok || img.Empty() {
                log.Println("Cannot read device or frame empty")
                continue
            }

            // Verify frame size
            if img.Cols() != 1280 || img.Rows() != 720 {
                log.Printf("Unexpected frame size: %dx%d", img.Cols(), img.Rows())
                continue
            }

            // Convert the frame to bytes
            frameBytes := img.ToBytes()
            expectedSize := 1280 * 720 * 3 // bgr24
            if len(frameBytes) != expectedSize {
                log.Printf("Unexpected frame byte size: got %d, expected %d", len(frameBytes), expectedSize)
                continue
            }

            // Write the frame to FFmpeg stdin
            _, err := stdin.Write(frameBytes)
            if err != nil {
                log.Printf("Failed to write frame to FFmpeg: %v", err)
                break
            }
        }
    }

    // Clean up (unreachable in this loop, but good practice)
    stdin.Close()
    ffmpegCmd.Wait()
}
