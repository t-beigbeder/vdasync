mkdir /local/tmp/copy-of-locgit
bin/vdasync -conc 4 -dryrun -rm -source ~/locgit -target /local/tmp/copy-of-locgit -silent
mkdir /local/tmp/copy-of-bin
bin/vdasync -conc 4 -dryrun -rm -source bin -target lf+dss:/local/tmp/copy-of-bin -config testdata/cmd/basicConfig.yaml -level INFO

time bin/vdasync -conc 4 -dryrun -rm -source ~/locgit -target lf+dss:/local/tmp/copy-of-locgit -config testdata/cmd/basicConfig.yaml -silent

mkdir /local/tmp/copy-of-home
ulimit -Sv 3000000

chmod -R +w /local/tmp/copy-of-locgit && rm -fr /local/tmp/copy-of-locgit && mkdir /local/tmp/copy-of-locgit

bin/vdasync -conc 4 -dryrun -rm -source ~/locgit -target /local/tmp/copy-of-locgit -silent -level INFO
bin/vdasync -conc 4 -rm -source ~/locgit -target /local/tmp/copy-of-locgit -config testdata/cmd/basicConfig.yaml -silent -level INFO
time=2026-05-12T16:46:34.914Z level=INFO msg="Run: root is done" app=vdasync walker=true "number processed"=11351 HeapInuse=25080 HeapAlloc=15465 StackInuse=7776

bin/vdasync -conc 4 -dryrun -rm -source ~/locgit -target lf+dss:/local/tmp/copy-of-locgit -config testdata/cmd/basicConfig.yaml -silent -level INFO
bin/vdasync -conc 4 -rm -source ~/locgit -target lf+dss:/local/tmp/copy-of-locgit -config testdata/cmd/basicConfig.yaml -silent -level INFO
time=2026-05-12T16:17:48.617Z level=INFO msg="Run: root is done" app=vdasync walker=true "number processed"=11315 HeapInuse=308912 HeapAlloc=268101 StackInuse=7488
time=2026-05-12T16:17:48.205Z level=INFO msg="RunOpeDssaServer: processed..." app=localFiles count=42000 HeapInuse=1104 HeapAlloc=604 StackInuse=384 statMap="map[List:1 Mkdir:2403 Put:8630 Put.Recv:8611 SetStat:11104 Stat:11187 Symlink:64]"

GOMEMLIMIT=250MiB bin/vdasync -conc 4 -rm -source ~/locgit -target lf+dss:/local/tmp/copy-of-locgit -config testdata/cmd/basicConfig.yaml -silent -level INFO
time=2026-05-12T16:53:27.543Z level=INFO msg="Run: root is done" app=vdasync walker=true "number processed"=11351 HeapInuse=184032 HeapAlloc=170153 StackInuse=4000


chmod -R +w /local/tmp/copy-of-home && rm -fr /local/tmp/copy-of-home && mkdir /local/tmp/copy-of-home

bin/vdasync -conc 4 -dryrun -rm -source ~ -target /local/tmp/copy-of-home -config testdata/cmd/basicConfig.yaml -silent -level INFO
bin/vdasync -conc 4 -rm -source ~ -target /local/tmp/copy-of-home -config testdata/cmd/basicConfig.yaml -silent -level INFO
time=2026-05-12T16:43:49.810Z level=INFO msg="Run: root is done" app=vdasync walker=true "number processed"=77629 HeapInuse=116224 HeapAlloc=91540 StackInuse=13440
> time=2026-05-12T19:27:54.914Z level=INFO msg="Run: root is done" app=vdasync walker=true count=77692 HeapInuse=85656 HeapAlloc=49855 StackInuse=2112


bin/vdasync -conc 4 -dryrun -rm -source ~ -target lf+dss:/local/tmp/copy-of-home -config testdata/cmd/basicConfig.yaml -silent -level INFO
bin/vdasync -conc 4 -rm -source ~ -target lf+dss:/local/tmp/copy-of-home -config testdata/cmd/basicConfig.yaml -silent -level INFO
time=2026-05-12T16:26:18.203Z level=INFO msg="Run: root is done" app=vdasync walker=true "number processed"=77621 HeapInuse=7356640 HeapAlloc=7340745 StackInuse=54688
time=2026-05-12T16:26:17.710Z level=INFO msg="RunOpeDssaServer: processed..." app=localFiles count=295000 HeapInuse=1072 HeapAlloc=598 StackInuse=352 statMap="map[List:1 Mkdir:12405 Put:63911 Put.Recv:63883 SetStat:77294 Stat:77442 Symlink:64]"
> time=2026-05-12T19:30:25.379Z level=INFO msg="Run: root is done" app=vdasync walker=true count=77693 HeapInuse=5936392 HeapAlloc=5693931 StackInuse=14848
> time=2026-05-12T19:30:25.221Z level=INFO msg="RunOpeDssaServer: processed..." app=localFiles count=296000 HeapInuse=1104 HeapAlloc=603 StackInuse=384 statMap="map[List:1 Mkdir:12414 Put:64124 Put.Recv:64123 SetStat:77633 Stat:77641 Symlink:64]"
>> time=2026-05-12T19:48:34.122Z level=INFO msg="Run: root is done" app=vdasync walker=true count=77699 HeapInuse=5265352 HeapAlloc=5250202 StackInuse=41440
>> time=2026-05-12T19:48:33.918Z level=INFO msg="RunOpeDssaServer: processed..." app=localFiles count=296000 HeapInuse=1104 HeapAlloc=603 StackInuse=384 statMap="map[List:1 Mkdir:12414 Put:64127 Put.Recv:64126 SetStat:77626 Stat:77642 Symlink:64]"
>>> time=2026-05-12T20:30:16.610Z level=INFO msg="Run: root is done" app=vdasync walker=true count=77716 HeapInuse=1755520 HeapAlloc=1548769 StackInuse=27200
>>> time=2026-05-12T20:30:16.274Z level=INFO msg="RunOpeDssaServer: processed..." app=localFiles count=296000 HeapInuse=1256 HeapAlloc=604 StackInuse=384 statMap="map[List:1 Mkdir:12414 Put:64128 Put.Recv:64126 SetStat:77624 Stat:77643 Symlink:64]"


GOMEMLIMIT=2GiB bin/vdasync -conc 4 -rm -source ~ -target lf+dss:/local/tmp/copy-of-home -config testdata/cmd/basicConfig.yaml -silent -level INFO
time=2026-05-12T16:57:27.425Z level=INFO msg="Run: processed..." app=vdasync walker=true "number processed"=66000 HeapInuse=6074008 HeapAlloc=6024994 StackInuse=177664
time=2026-05-12T16:57:33.164Z level=INFO msg="Run: root is done" app=vdasync walker=true "number processed"=77643 HeapInuse=1825056 HeapAlloc=1764423 StackInuse=23776
> time=2026-05-12T19:34:41.201Z level=INFO msg="Run: root is done" app=vdasync walker=true count=77695 HeapInuse=1897032 HeapAlloc=1830971 StackInuse=15456


walker_test
time=2026-05-12T17:20:16.583Z level=INFO msg="Run: root is done" walker=true count=108 HeapInuse=1816 HeapAlloc=766 StackInuse=704