build:
	go build .

cf:
	cf uninstall-plugin VerifyManifest || true
	cf install-plugin verify-manifest -f

it: build cf
