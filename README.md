
** DISCLAIMER **
this repo is not officially supported by tanzu.

## cf cli plugin for checking cf manifest.yml 
This plugin reads manifest.yml and checks follows:
- checking service instances specified from the manifest exists in the current target space
- checking routes specified from the manifest are available via [check-reserved-routes-for-a-domain cf api](https://v3-apidocs.cloudfoundry.org/version/3.197.0/index.html#check-reserved-routes-for-a-domain)

it exits with 0 if all good. exit 1 otherwise.

## Development
Clone the project and run following command. it builds plugin and install the built plugin into cf cli.
```sh
make it
```

## Build plugin per OS
plugin should be created under `./builds` folder.

#### For MacOS:
```sh
build-release-darwin
cf uninstall-plugin VerifyManifest
cf install-plugin ./builds/verify-manifest-darwin
```

#### For Linux
```sh
build-release
cf uninstall-plugin VerifyManifest
cf install-plugin ./builds/verify-manifest-linux
```
#### For Windows
```sh
build-release
cf uninstall-plugin VerifyManifest
cf install-plugin ./builds/verify-manifest.exe
```


list the installed plugins 
```sh
cf plugins
```

```sh
Listing installed plugins...

plugin                version   command name                  command help
VerifyManifest        1.0.0     verify-manifest               verify manifest.yml for service instances and routes

Use 'cf repo-plugins' to list plugins in registered repos available to install
```

## Executing the plugin

First, needs log in to the target foundation. otherwise it fails to `cf target` command.
```sh
cf login
```

### good case
Run the plugin with manifest.
```sh
cf  verify-manifest -f ./fixtures/manifest-good.yml
```
```
Using manifestPath: './fixtures/manifest-good.yml'
cf target ...
   API endpoint:   https://api.sys.dhaka.cf-app.com (API version: 3.194.0)
   User:           minseok.kim@broadcom.com
   Org:            minseok
   Space:          test

Checking Service instance from the manifest ... ./fixtures/manifest-good.yml
  Checking manifest app name ...
  Parsing manifest services ...
  Fetching cf services from the target foundation ...
  [INFO] Existing service instances in current space:
  - 'existing-service-1'
  - 'existing-service-2'
  - 'my-cups'
  - 'my-cups 3'
  - 'my-cups2'
  [GOOD] Service instance specified in manifest that exists in current space:
  - good app: 'spring-music', service: 'existing-service-1'
  - good app: 'spring-music2', service: 'existing-service-1'
  - good app: 'spring-music2', service: 'existing-service-2'

Checking Routes availability specified in the manifest from the target ...  ./fixtures/manifest-good.yml
  Checking manifest app name ...
  Parsing manifest routes ...
  Fetching cf domains, guid from the target foundation ...
  [GOOD] Available routes specified in the manifest in the target foundation:
  - good app: 'spring-music', route: 'spring-musicx.apps.internal'
  - good app: 'spring-music2', route: 'spring-music2.apps.internal'
```

return value from the plugin 
```
echo $?

0
```

### bad case.
```sh
cf  verify-manifest -f ./fixtures/manifest-bad.yml
```

```
Using manifestPath: './fixtures/manifest-bad.yml'
cf target ...
   API endpoint:   https://api.sys.dhaka.cf-app.com (API version: 3.194.0)
   User:           minseok.kim@broadcom.com
   Org:            minseok
   Space:          test

Checking Service instance from the manifest ... ./fixtures/manifest-bad.yml
  Checking manifest app name ...
  Parsing manifest services ...
  Fetching cf services from the target foundation ...
  [INFO] Existing service instances in current space:
  - 'existing-service-1'
  - 'existing-service-2'
  - 'my-cups'
  - 'my-cups 3'
  - 'my-cups2'
  [GOOD] Service instance specified in manifest that exists in current space:
  - good app: 'spring-music', service: 'existing-service-1'
  - good app: 'spring-music2', service: 'existing-service-2'
  - good app: 'spring-music2', service: 'existing-service-1'
  [ERROR] Missing Service instance specified in manifest but not exist in current space:
  - error app: 'spring-music', service: 'service-not-exist1'
  - error app: 'spring-music', service: 'service-not-exist2'
  - error app: 'spring-music2', service: '2service-not-exist1'
  - error app: 'spring-music2', service: '2service-not-exist2'

Checking Routes availability specified in the manifest from the target ...  ./fixtures/manifest-bad.yml
  Checking manifest app name ...
  Parsing manifest routes ...
  Fetching cf domains, guid from the target foundation ...
  [GOOD] Available routes specified in the manifest in the target foundation:
  - good app: 'spring-music', route: 'spring-music.apps.internal'
  - good app: 'spring-music2', route: 'spring-music2.apps.internal'
  [ERROR] Not Available routes specified in the manifest in the target foundation:
  - error app: 'spring-music', route: 'cryodocs.apps.dhaka.cf-app.com' -> Reserved is reserved
  - error app: 'spring-music', route: 'apps.internal' -> No such domain 'internal' in cf domains
  - error app: 'spring-music', route: 'internal' -> Invalid route. too short

$ echo $?
1
```

