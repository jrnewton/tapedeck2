# Please use Vim to edit me.  
# It knows to use tabs and ignore your editor settings.

app_name := tapedeck
out_dir := ./dist
bin_name := $(app_name)
bin_remote := /usr/local/$(app_name)
bin_local := $(out_dir)$(bin_remote)
config_remote := /etc/$(app_name)
config_local := $(out_dir)$(config_remote)
db_remote := /var/local/$(app_name)
db_local := $(out_dir)/$(db_remote)
user_remote := /var/local/$(app_name)/user
user_local := $(out_dir)$(user_remote)
web_remote := /usr/local/$(app_name)/web
web_local := $(out_dir)$(web_remote)

## package commands produce local tar.gz files
pkgupdate: vars clean out_dir binary web package
	@echo Packaging update...

pkginstall: pkgupdate db user config package
	@echo Packaging install...

package: vars clean
	tar -C $(out_dir) -czf /tmp/$(app_name).tar.gz .
	cp /tmp/$(app_name).tar.gz $(out_dir)

## prod commands update the production server
# ssh/rsync commands expect an entry in ~/.ssh/config for tapedeck root user
produpdate: vars clean out_dir binary web
	# Point 1 - do not use -a; my local user does not exist on remote & I want 
	#           everything as root.
	# Point 2 - trailing slash important! Copy _contents_ of left dir into right dir.
	rsync -rv --delete $(bin_local)/ tapedeck:$(bin_remote)
	rsync -rv --delete $(web_local)/ tapedeck:$(web_remote)

prodinstall: produpdate db user config
	rsync -rv --delete $(config_local)/ tapedeck:$(config_remote)
	rsync -rv --delete $(db_local)/ tapedeck:$(db_remote)
	rsync -rv --delete $(user_local)/ tapedeck:$(user_remote)

## reload the production service
reload: upload
	ssh tapedeck 'systemctl restart $(app_name)'

## the rest of the file
binary: vars out_dir
	@echo Creating binary dir...
	mkdir -p $(bin_local)
	@echo Building binary...
	go build -o $(bin_local)/$(bin_name) ./cmd/server/
	chmod +x $(bin_local)/$(bin_name)

config: vars out_dir
	@echo Creating config dir...
	mkdir -p $(config_local)
	cp ./config/prod/tapedeck.json $(config_local)

web: vars out_dir
	@echo Creating web dir...
	mkdir -p $(web_local)
	cp -a ./static/. $(web_local)/static/
	cp -a ./templates/. $(web_local)/templates/

user: vars out_dir
	@echo Creating user dir...
	mkdir -p $(user_local)

db: vars out_dir
	@echo Creating user dir...
	mkdir -p $(db_local)
	cp $(app_name).db $(db_local)

out_dir: vars
	@echo Creating output directory...
	mkdir -p $(out_dir)

vars:
	@echo -------------------------------------------------
	@echo Compiled binary is $(bin_name)
	@echo Output directory is $(out_dir)
	@echo binary dir is $(bin_local)
	@echo config dir is $(config_local)
	@echo db dir is $(db_local)
	@echo user dir is $(user_local)
	@echo web dir is $(web_local)
	@echo -------------------------------------------------

clean: vars
	echo Cleaning output directory...
	rm -rf $(out_dir)
