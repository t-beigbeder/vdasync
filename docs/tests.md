mkdir /local/tmp/copy-of-locgit
bin/vdasync -conc 4 -dryrun -rm -source ~/locgit -target /local/tmp/copy-of-locgit -silent
mkdir /local/tmp/copy-of-bin
bin/vdasync -conc 4 -dryrun -rm -source bin -target lf+dss:/local/tmp/copy-of-bin -config testdata/cmd/basicConfig.yaml -level INFO

time bin/vdasync -conc 4 -dryrun -rm -source ~/locgit -target lf+dss:/local/tmp/copy-of-locgit -config testdata/cmd/basicConfig.yaml -silent

mkdir /local/tmp/copy-of-home
ulimit -Sv 3000000

mkdir /local/tmp/copy-of-locgit
chmod -R +w /local/tmp/copy-of-locgit && rm -fr /local/tmp/copy-of-locgit && mkdir /local/tmp/copy-of-locgit
bin/vdasync -conc 4 -dryrun -rm -source ~/locgit -target /local/tmp/copy-of-locgit -silent -level INFO
bin/vdasync -conc 4 -rm -source ~/locgit -target /local/tmp/copy-of-locgit -config testdata/cmd/basicConfig.yaml -silent -level INFO

bin/vdasync -conc 4 -dryrun -rm -source ~/locgit -target lf+dss:/local/tmp/copy-of-locgit -config testdata/cmd/basicConfig.yaml -silent -level INFO
bin/vdasync -conc 4 -rm -source ~/locgit -target lf+dss:/local/tmp/copy-of-locgit -config testdata/cmd/basicConfig.yaml -silent -level INFO
time=2026-05-12T16:17:48.617Z level=INFO msg="Run: root is done" app=vdasync walker=true "number processed"=11315 HeapInuse=308912 HeapAlloc=268101 StackInuse=7488
time=2026-05-12T16:17:48.205Z level=INFO msg="RunOpeDssaServer: processed..." app=localFiles count=42000 HeapInuse=1104 HeapAlloc=604 StackInuse=384 statMap="map[List:1 Mkdir:2403 Put:8630 Put.Recv:8611 SetStat:11104 Stat:11187 Symlink:64]"

bin/vdasync -conc 4 -rm -source ~ -target /local/tmp/copy-of-home -config testdata/cmd/basicConfig.yaml -silent -level INFO

bin/vdasync -conc 4 -rm -source ~ -target /local/tmp/copy-of-home -config testdata/cmd/basicConfig.yaml -silent -level INFO
time=2026-05-12T16:43:49.810Z level=INFO msg="Run: root is done" app=vdasync walker=true "number processed"=77629 HeapInuse=116224 HeapAlloc=91540 StackInuse=13440

bin/vdasync -conc 4 -dryrun -rm -source ~ -target lf+dss:/local/tmp/copy-of-home -config testdata/cmd/basicConfig.yaml -silent -level INFO
bin/vdasync -conc 4 -rm -source ~ -target lf+dss:/local/tmp/copy-of-home -config testdata/cmd/basicConfig.yaml -silent -level INFO
time=2026-05-12T16:26:18.203Z level=INFO msg="Run: root is done" app=vdasync walker=true "number processed"=77621 HeapInuse=7356640 HeapAlloc=7340745 StackInuse=54688
time=2026-05-12T16:26:17.710Z level=INFO msg="RunOpeDssaServer: processed..." app=localFiles count=295000 HeapInuse=1072 HeapAlloc=598 StackInuse=352 statMap="map[List:1 Mkdir:12405 Put:63911 Put.Recv:63883 SetStat:77294 Stat:77442 Symlink:64]"

time bin/vdasync -conc 4 -rm -source ~ -target /local/tmp/copy-of-home -config testdata/cmd/basicConfig.yaml -silent
time=2026-05-10T17:20:15.314Z level=INFO msg="localFiles.main starting" app=localFiles host=localhost port=40517
time=2026-05-10T17:20:27.462Z level=INFO msg="localFiles.main done" app=localFiles host=localhost port=40517

real    0m7.675s
user    0m1.739s
sys     0m4.345s

chmod -R +w /local/tmp/copy-of-home && rm -fr /local/tmp/copy-of-home && mkdir /local/tmp/copy-of-home

time bin/vdasync -conc 4 -rm -source ~ -target lf+dss:/local/tmp/copy-of-home -config testdata/cmd/basicConfig.yaml -silent