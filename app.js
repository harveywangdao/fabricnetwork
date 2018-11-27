/**
 * Copyright 2017 IBM All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the 'License');
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an 'AS IS' BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */
'use strict';
var log4js = require('log4js');
var logger = log4js.getLogger('Fabric-Network');
logger.setLevel('INFO');

var uuid = require('uuid/v4');
var express = require('express');
var session = require('express-session');
var cookieParser = require('cookie-parser');
var bodyParser = require('body-parser');
var http = require('http');
var util = require('util');
var app = express();
var bearerToken = require('express-bearer-token');
var cors = require('cors');

require('./config.js');
var config = require('./config.json');

var hfc = require('fabric-client');

var helper = require('./app/helper.js');
var createChannel = require('./app/create-channel.js');
var join = require('./app/join-channel.js');
var install = require('./app/install-chaincode.js');
var instantiate = require('./app/instantiate-chaincode.js');
var invoke = require('./app/invoke-transaction.js');
var query = require('./app/query.js');
var host = process.env.HOST || hfc.getConfigSetting('host');
var port = process.env.PORT || hfc.getConfigSetting('port');
///////////////////////////////////////////////////////////////////////////////
//////////////////////////////// SET CONFIGURATONS ////////////////////////////
///////////////////////////////////////////////////////////////////////////////
app.options('*', cors());
app.use(cors());
//support parsing of application/json type post data
app.use(bodyParser.json());
//support parsing of application/x-www-form-urlencoded post data
app.use(bodyParser.urlencoded({
	extended: false
}));

app.use(bearerToken());

///////////////////////////////////////////////////////////////////////////////
//////////////////////////////// START SERVER /////////////////////////////////
///////////////////////////////////////////////////////////////////////////////
var server = http.createServer(app).listen(port, function() {});
logger.info('****************** SERVER STARTED ************************');
logger.info('***************  http://%s:%s  ******************',host,port);
server.timeout = 240000;

function getErrorMessage(field) {
	var response = {
		success: false,
		message: field + ' field is missing or Invalid in the request'
	};
	return response;
}

///////////////////////////////////////////////////////////////////////////////
///////////////////////// REST ENDPOINTS START HERE ///////////////////////////
///////////////////////////////////////////////////////////////////////////////
// Register and enroll user
app.post('/users', async function(req, res) {
	var username = req.body.username;
	var orgname = req.body.orgname;

	logger.debug('username :' + username);
	logger.debug('orgname:' + orgname);

	if (!username) {
		res.json(getErrorMessage('\'username\''));
		return;
	}

	if (!orgname) {
		res.json(getErrorMessage('\'orgname\''));
		return;
	}

	let response = await helper.getRegisteredUser(username, orgname, true);
	if (response && typeof response !== 'string') {
		logger.info('Success to register the username %s for organization %s with::%s', username, orgname, response);
		res.json(response);
	} else {
		logger.error('Failed to register the username %s for organization %s with::%s', username, orgname, response);
		res.json({success: false, message: response});
	}
});

// Create Channel
app.post('/channels', async function(req, res) {
	logger.info('<<<<<<<<<<<<<<<<<Create Channel>>>>>>>>>>>>>>>>>');
	var channelName = req.body.channelName;
	var channelConfigPath = req.body.channelConfigPath;
	var username = config.user1.username;
	var orgname = config.user1.org;

	logger.debug('Channel name : ' + channelName);
	logger.debug('channelConfigPath : ' + channelConfigPath); //../artifacts/channel/mychannel.tx

	if (!channelName) {
		res.json(getErrorMessage('\'channelName\''));
		return;
	}

	if (!channelConfigPath) {
		res.json(getErrorMessage('\'channelConfigPath\''));
		return;
	}

	let message = await createChannel.createChannel(channelName, channelConfigPath, username, orgname);
	res.send(message);
});

// Join Channel
app.post('/channels/:channelName/peers', async function(req, res) {
	logger.info('<<<<<<<<<<<<<<<<<Join Channel>>>>>>>>>>>>>>>>>');
	var channelName = req.params.channelName;
	var peers = req.body.peers;
	var username = req.body.username;
	var orgname = req.body.orgname;

	logger.debug('channelName : ' + channelName);
	logger.debug('peers : ' + peers);
	logger.debug('username :' + username);
	logger.debug('orgname:' + orgname);

	if (!channelName) {
		res.json(getErrorMessage('\'channelName\''));
		return;
	}

	if (!peers || peers.length == 0) {
		res.json(getErrorMessage('\'peers\''));
		return;
	}

	if (!username) {
		res.json(getErrorMessage('\'username\''));
		return;
	}

	if (!orgname) {
		res.json(getErrorMessage('\'orgname\''));
		return;
	}

	let message =  await join.joinChannel(channelName, peers, username, orgname);
	res.send(message);
});

// Install chaincode on target peers
app.post('/chaincodes', async function(req, res) {
	logger.info('==================== INSTALL CHAINCODE ==================');
	var peers = req.body.peers;
	var chaincodeName = req.body.chaincodeName;
	var chaincodePath = req.body.chaincodePath;
	var chaincodeVersion = req.body.chaincodeVersion;
	var chaincodeType = req.body.chaincodeType;
	var username = req.body.username;
	var orgname = req.body.orgname;

	logger.debug('peers : ' + peers); // target peers list
	logger.debug('chaincodeName : ' + chaincodeName);
	logger.debug('chaincodePath  : ' + chaincodePath);
	logger.debug('chaincodeVersion  : ' + chaincodeVersion);
	logger.debug('chaincodeType  : ' + chaincodeType);
	logger.debug('username :' + username);
	logger.debug('orgname:' + orgname);

	if (!peers || peers.length == 0) {
		res.json(getErrorMessage('\'peers\''));
		return;
	}

	if (!chaincodeName) {
		res.json(getErrorMessage('\'chaincodeName\''));
		return;
	}

	if (!chaincodePath) {
		res.json(getErrorMessage('\'chaincodePath\''));
		return;
	}

	if (!chaincodeVersion) {
		res.json(getErrorMessage('\'chaincodeVersion\''));
		return;
	}

	if (!chaincodeType) {
		res.json(getErrorMessage('\'chaincodeType\''));
		return;
	}

	if (!username) {
		res.json(getErrorMessage('\'username\''));
		return;
	}

	if (!orgname) {
		res.json(getErrorMessage('\'orgname\''));
		return;
	}

	let message = await install.installChaincode(peers, chaincodeName, chaincodePath, chaincodeVersion, chaincodeType, username, orgname)
	res.send(message);
});

// Instantiate chaincode on target peers
app.post('/channels/:channelName/chaincodes', async function(req, res) {
	logger.info('==================== INSTANTIATE CHAINCODE ==================');
	var peers = req.body.peers;
	var chaincodeName = req.body.chaincodeName;
	var chaincodeVersion = req.body.chaincodeVersion;
	var channelName = req.params.channelName;
	var chaincodeType = req.body.chaincodeType;
	var fcn = req.body.fcn;
	var args = req.body.args;
	var username = req.body.username;
	var orgname = req.body.orgname;

	logger.debug('peers  : ' + peers);
	logger.debug('channelName  : ' + channelName);
	logger.debug('chaincodeName : ' + chaincodeName);
	logger.debug('chaincodeVersion  : ' + chaincodeVersion);
	logger.debug('chaincodeType  : ' + chaincodeType);
	logger.debug('fcn  : ' + fcn);
	logger.debug('args  : ' + args);
	logger.debug('username :' + username);
	logger.debug('orgname:' + orgname);

	if (!chaincodeName) {
		res.json(getErrorMessage('\'chaincodeName\''));
		return;
	}

	if (!chaincodeVersion) {
		res.json(getErrorMessage('\'chaincodeVersion\''));
		return;
	}

	if (!channelName) {
		res.json(getErrorMessage('\'channelName\''));
		return;
	}

	if (!chaincodeType) {
		res.json(getErrorMessage('\'chaincodeType\''));
		return;
	}

	if (!args) {
		res.json(getErrorMessage('\'args\''));
		return;
	}

	if (!username) {
		res.json(getErrorMessage('\'username\''));
		return;
	}

	if (!orgname) {
		res.json(getErrorMessage('\'orgname\''));
		return;
	}

	let message = await instantiate.instantiateChaincode(peers, channelName, chaincodeName, chaincodeVersion, chaincodeType, fcn, args, username, orgname);
	res.send(message);
});

// Invoke transaction on chaincode on target peers
app.post('/channels/:channelName/chaincodes/:chaincodeName', async function(req, res) {
	logger.info('==================== INVOKE ON CHAINCODE ==================');
	var peers = req.body.peers;
	var chaincodeName = req.params.chaincodeName;
	var channelName = req.params.channelName;
	var fcn = req.body.fcn;
	var args = req.body.args;
	var username = config.user1.username;
	var orgname = config.user1.org;

	logger.debug('channelName  : ' + channelName);
	logger.debug('chaincodeName : ' + chaincodeName);
	logger.debug('fcn  : ' + fcn);
	logger.debug('args  : ' + args);

	if (!chaincodeName) {
		res.json(getErrorMessage('\'chaincodeName\''));
		return;
	}

	if (!channelName) {
		res.json(getErrorMessage('\'channelName\''));
		return;
	}

	if (!fcn) {
		res.json(getErrorMessage('\'fcn\''));
		return;
	}

	if (!args) {
		res.json(getErrorMessage('\'args\''));
		return;
	}

	let message = await invoke.invokeChaincode(peers, channelName, chaincodeName, fcn, args, username, orgname);
	res.send(message);
});

app.post('/ocean/v1/issueToken', async function(req, res) {
	logger.info('==================== IssueToken ==================');
	var peers = config.invokePeers;
	var channelName = config.channelName;
	var chaincodeName = config.chaincodeName;

	var fcn = "issueToken";
	var pubKey = req.body.pubKey;
	var origin = req.body.origin;
	var signature = req.body.signature;

	var username = config.user1.username;
	var orgname = config.user1.org;

	if (!pubKey) {
		res.json(getErrorMessage('\'pubKey\''));
		return;
	}

	if (!origin) {
		res.json(getErrorMessage('\'origin\''));
		return;
	}

	if (!signature) {
		res.json(getErrorMessage('\'signature\''));
		return;
	}

	let tokenID = uuid().replace(/-/g, '');
	logger.info("tokenID =", tokenID);

	var args = [];
	args.push(tokenID);
	args.push(pubKey);
	args.push(origin);
	args.push(signature);

	let message = await invoke.invokeChaincode(peers, channelName, chaincodeName, fcn, args, username, orgname);
	message.tokenID = tokenID;
	res.send(message);
});

app.get('/ocean/v1/queryToken/:tokenID', async function(req, res) {
	logger.info('==================== QueryToken ==================');
	let tokenID = req.params.tokenID;

	let channelName = config.channelName;
	let chaincodeName = config.chaincodeName;
	let peer = config.queryPeers;
	let username = config.user2.username;
	let orgname = config.user2.org;
	let fcn = "queryToken";

	logger.info('tokenID : ' + tokenID);

	if (!tokenID) {
		res.json(getErrorMessage('\'tokenID\''));
		return;
	}

	let args = [];
	args.push(tokenID)

	let message = await query.queryChaincode(peer, channelName, chaincodeName, args, fcn, username, orgname);
	res.send(message);
});

app.get('/ocean/v1/queryBalance/:address', async function(req, res) {
	logger.info('==================== QueryBalance ==================');
	let address = req.params.address;

	let channelName = config.channelName;
	let chaincodeName = config.chaincodeName;
	let peer = config.queryPeers;
	let username = config.user2.username;
	let orgname = config.user2.org;
	let fcn = "queryBalance";

	logger.info('address : ' + address);

	if (!address) {
		res.json(getErrorMessage('\'address\''));
		return;
	}

	let args = [];
	args.push(address)

	let message = await query.queryChaincode(peer, channelName, chaincodeName, args, fcn, username, orgname);
	res.send(message);
});

// Query on chaincode on target peers
app.get('/channels/:channelName/chaincodes/:chaincodeName', async function(req, res) {
	logger.info('==================== QUERY BY CHAINCODE ==================');
	var channelName = req.params.channelName;
	var chaincodeName = req.params.chaincodeName;
	let args = req.query.args;
	let fcn = req.query.fcn;
	let peer = req.query.peer;
	var username = config.user2.username;
	var orgname = config.user2.org;

	logger.debug('channelName : ' + channelName);
	logger.debug('chaincodeName : ' + chaincodeName);
	logger.debug('fcn : ' + fcn);
	logger.debug('args : ' + args);

	if (!chaincodeName) {
		res.json(getErrorMessage('\'chaincodeName\''));
		return;
	}
	if (!channelName) {
		res.json(getErrorMessage('\'channelName\''));
		return;
	}
	if (!fcn) {
		res.json(getErrorMessage('\'fcn\''));
		return;
	}
	if (!args) {
		res.json(getErrorMessage('\'args\''));
		return;
	}
	args = args.replace(/'/g, '"');
	args = JSON.parse(args);
	logger.debug(args);

	let message = await query.queryChaincode(peer, channelName, chaincodeName, args, fcn, username, orgname);
	res.send(message);
});

//  Query Get Block by BlockNumber
app.get('/channels/:channelName/blocks/:blockId', async function(req, res) {
	logger.debug('==================== GET BLOCK BY NUMBER ==================');
	let blockId = req.params.blockId;
	let peer = req.query.peer;
	var username = config.user2.username;
	var orgname = config.user2.org;

	logger.debug('channelName : ' + req.params.channelName);
	logger.debug('BlockID : ' + blockId);
	logger.debug('Peer : ' + peer);

	if (!blockId) {
		res.json(getErrorMessage('\'blockId\''));
		return;
	}

	let message = await query.getBlockByNumber(peer, req.params.channelName, blockId, username, orgname);
	res.send(message);
});

// Query Get Transaction by Transaction ID
app.get('/channels/:channelName/transactions/:trxnId', async function(req, res) {
	logger.debug('================ GET TRANSACTION BY TRANSACTION_ID ======================');
	logger.debug('channelName : ' + req.params.channelName);
	let trxnId = req.params.trxnId;
	let peer = req.query.peer;
	var username = config.user2.username;
	var orgname = config.user2.org;

	if (!trxnId) {
		res.json(getErrorMessage('\'trxnId\''));
		return;
	}

	let message = await query.getTransactionByID(peer, req.params.channelName, trxnId, username, orgname);
	res.send(message);
});

// Query Get Block by Hash
app.get('/channels/:channelName/blocks', async function(req, res) {
	logger.debug('================ GET BLOCK BY HASH ======================');
	logger.debug('channelName : ' + req.params.channelName);
	let hash = req.query.hash;
	let peer = req.query.peer;
	var username = config.user2.username;
	var orgname = config.user2.org;

	if (!hash) {
		res.json(getErrorMessage('\'hash\''));
		return;
	}

	let message = await query.getBlockByHash(peer, req.params.channelName, hash, username, orgname);
	res.send(message);
});

//Query for Channel Information
app.get('/channels/:channelName', async function(req, res) {
	logger.debug('================ GET CHANNEL INFORMATION ======================');
	logger.debug('channelName : ' + req.params.channelName);
	let peer = req.query.peer;
	var username = config.user2.username;
	var orgname = config.user2.org;

	let message = await query.getChainInfo(peer, req.params.channelName, username, orgname);
	res.send(message);
});

//Query for Channel instantiated chaincodes
app.get('/channels/:channelName/chaincodes', async function(req, res) {
	logger.debug('================ GET INSTANTIATED CHAINCODES ======================');
	logger.debug('channelName : ' + req.params.channelName);
	let peer = req.query.peer;
	var username = config.user2.username;
	var orgname = config.user2.org;

	let message = await query.getInstalledChaincodes(peer, req.params.channelName, 'instantiated', username, orgname);
	res.send(message);
});

// Query to fetch all Installed/instantiated chaincodes
app.get('/chaincodes', async function(req, res) {
	logger.debug('================ GET INSTALLED CHAINCODES ======================');
	var peer = req.query.peer;
	var installType = req.query.type;
	var username = config.user2.username;
	var orgname = config.user2.org;

	let message = await query.getInstalledChaincodes(peer, null, 'installed', username, orgname)
	res.send(message);
});

// Query to fetch channels
app.get('/channels', async function(req, res) {
	logger.debug('================ GET CHANNELS ======================');
	logger.debug('peer: ' + req.query.peer);
	var peer = req.query.peer;
	var username = config.user2.username;
	var orgname = config.user2.org;

	if (!peer) {
		res.json(getErrorMessage('\'peer\''));
		return;
	}

	let message = await query.getChannels(peer, username, orgname);
	res.send(message);
});
