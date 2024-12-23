# Please use Vim to edit me.  
# It knows to use tabs and ignore your editor settings.

out_dir := ./build_output
server_root := $(out_dir)/server
bin_name := tapedeck

package: vars binary out_dir server_root
	@echo Creating distribution package...
	cp run-dev.sh $(out_dir)
	cp run-prod.sh $(out_dir)
	chmod +x $(out_dir)/run-dev.sh
	chmod +x $(out_dir)/run-prod.sh
	cp -r ./static/ $(server_root)/static 
	cp -r ./templates/ $(server_root)/templates 
	cp config-prod.json $(out_dir)
	tar -C $(out_dir) -czf $(bin_name).tar.gz .

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
	scp $(bin_name).tar.gz tapedeck:/usr/local/tapedeck
