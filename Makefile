build:

	go build -o ./builds/verify-manifest .

cf:
	cf uninstall-plugin VerifyManifest || true
	cf install-plugin ./builds/verify-manifest -f

all: build cf
