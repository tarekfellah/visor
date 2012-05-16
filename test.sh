set -x

export VISOR_DEBUG=1
export VISOR_ROOT=/bazooka

visor=./bin/visor

$visor root init

$visor app register -t lxc -u http://github.com/foo/bar bar

$visor revision register bar cb876c2 -u bazooka@foo.bar.baz:/srv/files/bar_cb876c2.img -t web
$visor revision register bar 5ff435f -u bazooka@foo.bar.baz:/srv/files/bar_5ff435f.img -t web
$visor revision register bar 35f2c83 -u bazooka@foo.bar.baz:/srv/files/bar_35f2c83.img -t web

$visor instance create bar 35f2c83 web 10.20.30.40:20000
$visor instance create bar 35f2c83 web 10.20.30.41:20000
$visor instance create bar 35f2c83 web 10.20.30.42:20000
$visor instance create bar 35f2c83 web 10.20.30.43:20000

#####################################################################################################

$visor app register -t lxc -u http://github.com/foo/baz baz

$visor revision register baz 2fe0376 -u bazooka@foo.bar.baz:/srv/files/baz_2fe0376.img -t proc2,proc1
$visor revision register baz d228b00 -u bazooka@foo.bar.baz:/srv/files/baz_d228b00.img -t proc2,proc1
$visor revision register baz c16a2c5 -u bazooka@foo.bar.baz:/srv/files/baz_c16a2c5.img -t proc2,proc1

$visor instance create baz 2fe0376 proc1 10.20.30.40:20001
$visor instance create baz 2fe0376 proc2 10.20.30.40:20002
$visor instance create baz 2fe0376 proc1 10.20.30.41:20003
$visor instance create baz 2fe0376 proc2 10.20.30.41:20004

$visor ticket create baz 2fe0376 proc1 start
$visor ticket create baz 2fe0376 proc1 start
$visor ticket create baz 2fe0376 proc1 start
$visor ticket create baz 2fe0376 proc1 start
$visor ticket create baz 2fe0376 proc1 start

$visor revision scale baz 2fe0376 proc2 10
$visor revision scale baz 2fe0376 proc2 5

#####################################################################################################

$visor app register -t bazapta -u http://github.com/foo/kaboom kaboom

$visor revision register kaboom 4039763 -u bazooka@foo.bar.baz:/srv/files/kaboom_4039763.img
$visor revision register kaboom 0acdc3c -u bazooka@foo.bar.baz:/srv/files/kaboom_0acdc3c.img
$visor revision register kaboom 31cdec5 -u bazooka@foo.bar.baz:/srv/files/kaboom_31cdec5.img
