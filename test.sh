# Copyright (c) 2012, SoundCloud Ltd.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.
# Source code and contact info at http://github.com/soundcloud/visor

set -x

visor=./bin/visor
root=/bazooka

$visor -root=$root init

$visor -root=$root app-register bar
$visor -root=$root app-env-set bar KEY0 VALUE0
$visor -root=$root app-env-set bar KEY1 VALUE1
$visor -root=$root rev-register bar cb876c2 http://foo.bar.baz/bar_cb876c2.img
$visor -root=$root proc-register bar proc1

$visor -root=$root scale bar proc1 cb876c2 5

$visor -root=$root app-env-get bar
$visor -root=$root app-env-get bar KEY1
$visor -root=$root app-env-del bar KEY1
$visor -root=$root rev-exists bar cb876c2
$visor -root=$root rev-describe bar cb876c2
$visor -root=$root app-describe bar

$visor -root=$root proc-unregister bar proc1
$visor -root=$root rev-unregister bar cb876c2
$visor -root=$root app-unregister bar
