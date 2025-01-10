# Please use Vim to edit me.  
# It knows to use tabs and ignore your editor settings.

app_name := tapedeck
out_dir := ./dist
server_name := $(app_name)
app_dir_path := /usr/local/tapedeck
app_dir := $(out_dir)$(app_dir_path)
config_dir_path := /etc/tapedeck
config_dir := $(out_dir)$(config_dir_path)
bin_name := $(app_name)

package: collect
	@echo Creating distribution package...
	tar -C $(out_dir) -czf $(out_dir)/$(app_name).tar.gz .

collect: vars clean binary out_dir app_dir config_dir
	@echo Collecting distribution files...
	cp tapedeck.db $(app_dir)
	cp -a ./static/. $(app_dir)/static/
	cp -a ./templates/. $(app_dir)/templates/
	cp ./config/tapedeck.json $(config_dir)

binary: vars app_dir
	@echo Building binary...
	go build -o $(app_dir)/$(bin_name) ./cmd/
	chmod +x $(app_dir)/$(bin_name)

config_dir: vars out_dir
	@echo Creating config dir...
	mkdir -p $(config_dir)

app_dir: vars out_dir
	@echo Creating app dir...
	mkdir -p $(app_dir)

out_dir: vars
	@echo Creating output directory...
	mkdir -p $(out_dir)

vars:
	@echo -------------------------------------------------
	@echo Compiled binary is $(bin_name)
	@echo Output directory is $(out_dir)
	@echo app dir is $(app_dir)
	@echo config dir is $(config_dir)
	@echo -------------------------------------------------

clean: vars
	echo Cleaning output directory...
	rm -rf $(out_dir)

# ssh/rsync commands expect an entry in ~/.ssh/config for tapedeck root user

upload: collect
	# Point 1 - do not use -a; my local user does not exist on remote & I want 
	#           everything as root.
	# Point 2 - trailing slash important! Copy _contents_ of left dir into right dir.
	rsync -rv --delete $(app_dir)/ tapedeck:$(app_dir_path)
	rsync -rv --delete $(config_dir)/ tapedeck:$(config_dir_path)

reload: upload
	ssh tapedeck 'systemctl restart tapedeck'
