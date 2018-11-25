'use strict';
var log4js = require('log4js');
var logger = log4js.getLogger('Fabric-Network');
logger.setLevel('INFO');

var uuidv1 = require('uuid/v1');
var uuidv4 = require('uuid/v4');

logger.info(uuidv1().replace(/-/g, ''));
logger.info(uuidv1());
logger.info(uuidv1());
logger.info(uuidv1());
logger.info(uuidv1());

logger.info(uuidv4());
logger.info(uuidv4());
logger.info(uuidv4());
logger.info(uuidv4());
logger.info(uuidv4());