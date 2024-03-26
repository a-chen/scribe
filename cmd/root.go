/*
Copyright © 2024 Andrew Chen <achen.this@gmail.com>
*/
package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "scribe",
	Short: "CLI app that takes in audio/video files, YouTube links, and produces transcribed text",
	Long: `CLI app that takes in audio/video files, YouTube links, and produces transcribed text
Depending on your ASR source, transcription will take a couple of minutes

USAGE
scribe someYouTubeLink | output.txt 
scribe someAudioFile.mp3
scribe someVideoFile.mp4`,
	Args: cobra.ExactArgs(1),
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		var filePath = args[0]
		var audioFilePath = "tmp.mp3"
		// cleanup tmp file
		os.Remove(audioFilePath)

		//log.Printf("Processing %s\n", filePath)

		if strings.Contains(filePath, "youtube.com") {
			// extract audio file
			//log.Println("Extracting audio file...")
			cmd := exec.Command("yt-dlp", "--extract-audio", "--audio-format", "mp3", "--audio-quality", "10", "-o", audioFilePath, filePath)
			err := cmd.Run()
			if err != nil {
				log.Fatalf("Command execution failed with error: %v", err)
			}
		} else if strings.HasSuffix(filePath, ".mp3") {
			// feed audio file in
			audioFilePath = filePath
		} else {
			//log.Println("Extracting audio file...")
			cmd := exec.Command("ffmpeg", "-i", filePath, "-vn", "-acodec", "libmp3lame", audioFilePath)

			err := cmd.Run()
			if err != nil {
				log.Fatalf("Command execution failed with error: %v", err)
			}
		}

		// call WhisperASR
		// Prepare a form that you will submit via that POST.
		buf := new(bytes.Buffer)
		writer := multipart.NewWriter(buf)
		formFile, err := writer.CreateFormFile("audio_file", audioFilePath)
		if err != nil {
			log.Fatalf("Create form file failed: %s\n", err)
		}

		// Open the file
		fh, err := os.Open(audioFilePath)
		if err != nil {
			log.Fatalf("Error opening file: %v", err)
		}
		defer fh.Close()

		_, err = io.Copy(formFile, fh)
		if err != nil {
			log.Fatalf("Error copying data to form file: %v", err)
		}

		// Close multipart writer.
		writer.Close()

		// Create a client
		client := &http.Client{}

		var asrServerUrl = "http://0.0.0.0:9000"
		// Create a New Request
		// see https://ahmetoner.com/whisper-asr-webservice/endpoints/#automatic-speech-recognition-service-asr
		req, err := http.NewRequest(
			"POST",
			asrServerUrl+"/asr",
			buf,
		)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		// Do the request
		//log.Println("Transcribing...")
		res, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()

		//log.Println("Transcription complete")
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		fmt.Print(bodyString)
		// cleanup tmp file
		os.Remove(audioFilePath)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.scribe.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".scribe" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".scribe")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
