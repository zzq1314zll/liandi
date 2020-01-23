// LianDi - 链滴笔记，链接点滴
// Copyright (c) 2020-present, b3log.org
//
// Lute is licensed under the Mulan PSL v1.
// You can use this software according to the terms and conditions of the Mulan PSL v1.
// You may obtain a copy of Mulan PSL v1 at:
//     http://license.coscl.org.cn/MulanPSL
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v1 for more details.

package command

import (
	"os"

	"github.com/88250/gulu"
)

var logger = gulu.Log.NewLogger(os.Stdout)

type Cmd interface {
	Name() string
	Exec(map[string]interface{})
}

var Commands = map[string]Cmd{}

var (
	lsCmd = &ls{}
)

func init() {
	Commands[lsCmd.Name()] = lsCmd
}
