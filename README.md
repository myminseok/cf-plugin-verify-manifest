
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
  Parsing manifest services ...
  Fetching cf services from the target foundation ...
  [INFO] Existing service instances in current space:
  - 'my-cups'
  - 'my-cups 3'
  - 'my-cups2'
  [GOOD] Service instance specified in manifest exists in current space:
  - app: 'spring-music', service: 'my-cups2'
  - app: 'spring-music2', service: 'my-cups 3'
  - app: 'spring-music2', service: 'my-cups'

Checking Routes availability specified in the manifest from the target ...  ./fixtures/manifest-good.yml
  Parsing manifest routes ...
  Fetching cf domains, guid from the target foundation ...
  [GOOD] Available routes specified in the manifest in the target foundation:
  - app: 'spring-music', service: 'spring-musicx.apps.internal'
  - app: 'spring-music2', service: 'spring-music2.apps.internal'
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
  Parsing manifest services ...
  Fetching cf services from the target foundation ...
  [INFO] Existing service instances in current space:
  - 'my-cups'
  - 'my-cups 3'
  - 'my-cups2'
  [GOOD] Service instance specified in manifest exists in current space:
  - app: 'spring-music', service: 'my-cups2'
  - app: 'spring-music2', service: 'my-cups 3'
  - app: 'spring-music2', service: 'my-cups'
  [ERROR] Missing Service instance specified in manifest exists in current space:
  - app: 'spring-music', service: 'service-not-exist1'
  - app: 'spring-music', service: 'service-not-exist2'
  - app: 'spring-music2', service: '2service-not-exist1'
  - app: 'spring-music2', service: '2service-not-exist2'

Checking Routes availability specified in the manifest from the target ...  ./fixtures/manifest-bad.yml
  Parsing manifest routes ...
  Fetching cf domains, guid from the target foundation ...
  [GOOD] Available routes specified in the manifest in the target foundation:
  - app: 'spring-music', service: 'spring-music.apps.internal'
  - app: 'spring-music2', service: 'spring-music2.apps.internal'
  [ERROR] Not Available routes specified in the manifest in the target foundation:
  - app: 'spring-music', service: 'cryodocs.apps.dhaka.cf-app.com' -> Reserved is reserved
  - app: 'spring-music', service: 'apps.internal' -> No such domain 'internal' in cf domains
  - app: 'spring-music', service: 'internal' -> Invalid route. too short

$ echo $?
1
```

