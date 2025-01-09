# Please use Vim to edit me.  
# It knows to use tabs and ignore your editor settings.

app_name := tapedeck
out_dir := ./build_output
server_name := $(app_name)
server_root := $(out_dir)/server
bin_name := $(app_name)

package: collect
	@echo Creating distribution package...
	tar -C $(out_dir) -czf $(app_name).tar.gz .

collect: vars clean binary out_dir server_root
	@echo Collecting distribution files...
	cp start-dev.sh $(out_dir)
	cp start-prod.sh $(out_dir)
	chmod +x $(out_dir)/start-dev.sh
	chmod +x $(out_dir)/start-prod.sh
	cp tapedeck.db $(server_root)
	cp -a ./static/. $(server_root)/static/
	cp -a ./templates/. $(server_root)/templates/
	cp config-prod.json $(out_dir)

binary: vars out_dir
	@echo Building binary...
	go build -o $(out_dir)/$(bin_name) ./cmd/
	chmod +x $(out_dir)/$(bin_name)

server_root: vars out_dir
	@echo Creating server root...
	mkdir -p $(server_root)

out_dir: vars
	@echo Creating output directory...
	mkdir -p $(out_dir)

vars:
	@echo -------------------------------------------------
	@echo Compiled binary is $(bin_name)
	@echo Output directory is $(out_dir)
	@echo Server root directory is $(server_root)
	@echo -------------------------------------------------

clean: vars
	echo Cleaning output directory...
	rm -rf $(out_dir)

upload: package
	scp $(bin_name).tar.gz $(server_name):/usr/local/$(app_name)
