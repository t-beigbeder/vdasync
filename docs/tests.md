mkdir /local/tmp/copy-of-locgit
bin/vdasync -conc 4 -dryrun -rm -source ~/locgit -target /local/tmp/copy-of-locgit -silent
mkdir /local/tmp/copy-of-bin
bin/vdasync -conc 4 -dryrun -rm -source bin -target lf+dss:/local/tmp/copy-of-bin -config testdata/cmd/basicConfig.yaml -level INFO

time bin/vdasync -conc 4 -dryrun -rm -source ~/locgit -target lf+dss:/local/tmp/copy-of-locgit -config testdata/cmd/basicConfig.yaml -silent

mkdir /local/tmp/copy-of-home
ulimit -Sv 3000000

mkdir /local/tmp/copy-of-locgit
chmod -R +w /local/tmp/copy-of-locgit && rm -fr /local/tmp/copy-of-locgit && mkdir /local/tmp/copy-of-locgit
bin/vdasync -conc 0 -dryrun -rm -source ~/locgit -target /local/tmp/copy-of-locgit -config testdata/cmd/basicConfig.yaml -silent
bin/vdasync -conc 4 -dryrun -rm -source ~/locgit -target lf+dss:/local/tmp/copy-of-locgit -config testdata/cmd/basicConfig.yaml -silent

time bin/vdasync -conc 4 -dryrun -rm -source ~ -target lf+dss:/local/tmp/copy-of-home -config testdata/cmd/basicConfig.yaml -silent

time bin/vdasync -conc 4 -rm -source ~ -target /local/tmp/copy-of-home -config testdata/cmd/basicConfig.yaml -silent
time=2026-05-10T17:20:15.314Z level=INFO msg="localFiles.main starting" app=localFiles host=localhost port=40517
time=2026-05-10T17:20:27.462Z level=INFO msg="localFiles.main done" app=localFiles host=localhost port=40517

real    0m7.675s
user    0m1.739s
sys     0m4.345s

chmod -R +w /local/tmp/copy-of-home && rm -fr /local/tmp/copy-of-home && mkdir /local/tmp/copy-of-home

time bin/vdasync -conc 4 -rm -source ~ -target lf+dss:/local/tmp/copy-of-home -config testdata/cmd/basicConfig.yaml -silent