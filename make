mkdir static
mkdir release
cd static
EXT="session,curl,phar,posix,readline,fileinfo,filter,tokenizer,pcntl,bz2,bcmath,openssl,ftp,mbstring"
spc download --with-php=8.3 --all
spc build --build-cli --build-micro "$EXT"
spc micro:combine ../builds/efa -O ../release/efa
