package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gocv.io/x/gocv"
)

func main() {
    // RTSP server URL
    rtspURL := "rtsp://192.168.1.86:8554/mystream" // Replace with your server's IP and stream name

    // Open the RTSP stream
    stream, err := gocv.VideoCaptureFile(rtspURL)
    if err != nil {
        log.Fatalf("Error opening video capture: %v", err)
    }
    defer stream.Close()

    // Verify if the stream is opened
    if !stream.IsOpened() {
        log.Fatalf("Failed to open the RTSP stream: %s", rtspURL)
    }
    log.Println("Successfully connected to the RTSP stream")

    // Create a window to display the stream
    window := gocv.NewWindow("RTSP Receiver")
    defer window.Close()

    // Create a Mat to hold frames
    frame := gocv.NewMat()
    defer frame.Close()

    // Set up signal handling to gracefully exit
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

    // Start time for FPS calculation
    startTime := time.Now()
    frameCount := 0

    // Main loop to read frames
    for {
        if ok := stream.Read(&frame); !ok {
            log.Println("Cannot read frame from RTSP stream")
            continue
        }
        if frame.Empty() {
            continue
        }

        // Display the frame in the window
        window.IMShow(frame)

        // Increment frame count
        frameCount++

        // Calculate and display FPS every second
        elapsed := time.Since(startTime)
        if elapsed >= time.Second {
            fps := float64(frameCount) / elapsed.Seconds()
            window.SetWindowTitle(fmt.Sprintf("RTSP Receiver - FPS: %.2f", fps))
            frameCount = 0
            startTime = time.Now()
        }

        // Check if a key was pressed (e.g., to close the window)
        if window.WaitKey(1) >= 0 {
            log.Println("Key pressed. Exiting...")
            break
        }

        // Handle signal-based exit
        select {
        case <-sigs:
            log.Println("Interrupt received. Exiting...")
            return
        default:
        }
    }
}
