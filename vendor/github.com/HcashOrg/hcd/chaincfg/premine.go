// Copyright (c) 2014 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers 
// Copyright (c) 2018-2020 The Hc developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chaincfg

// BlockOneLedgerMainNet is the block one output ledger for the main
// network.
var BlockOneLedgerMainNet = []*TokenPayout{
	{"HsDMnqM8qaDquPeqYG9XUkRb9CWBWcevJue",40000000 * 1e8},
	{"TspAtkQvUoYKMUzxtTQSUuk6gECEWd7ZuUV",2000000 * 1e8},
	{"TbQ6WRtrx8w9oWXf49vyTqL3xDdsNRy1m7q",1326010 * 1e8},
}

// BlockOneLedgerTestNet is the block one output ledger for the test
// network.
var BlockOneLedgerTestNet = []*TokenPayout{
	{"TspAtg3jUAieeHhR9wgWso5CET3tZxQmvLq", 40000000 * 1e8},
}

// BlockOneLedgerTestNet2 is the block one output ledger for the 2nd test
// network.
var BlockOneLedgerTestNet2 = []*TokenPayout{
	{"TspAtg3jUAieeHhR9wgWso5CET3tZxQmvLq", 40000000 * 1e8},
}

// BlockOneLedgerSimNet is the block one output ledger for the simulation
// network. See under "Hcd organization related parameters" in params.go
// for information on how to spend these outputs.
var BlockOneLedgerSimNet = []*TokenPayout{
	{"Sshw6S86G2bV6W32cbc7EhtFy8f93rU6pae", 100000 * 1e8},
	{"SsjXRK6Xz6CFuBt6PugBvrkdAa4xGbcZ18w", 100000 * 1e8},
	{"SsfXiYkYkCoo31CuVQw428N6wWKus2ZEw5X", 100000 * 1e8},
}
