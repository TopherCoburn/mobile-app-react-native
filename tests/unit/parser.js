import { config } from './config';
import { parseString } from 'xml2js';
import { parse } from 'json-stream';

const parseXml = (xmlString) => {
  return new Promise((resolve, reject) => {
    parseString(xmlString, (err, result) => {
      if (err) {
        reject(err);
      } else {
        resolve(result);
      }
    });
  });
};

const parseJsonStream = (jsonStream) => {
  let result = [];

  jsonStream.on('data', (chunk) => {
    result.push(chunk);
  });

  jsonStream.on('end', () => {
    resolve(result);
  });
};

const parseFile = (filePath) => {
  return new Promise((resolve, reject) => {
    const fs = require('fs');
    const fileStream = fs.createReadStream(filePath);

    const parser = parse();
    fileStream.pipe(parser);

    fileStream.on('error', (err) => {
      reject(err);
    });

    parser.on('data', (chunk) => {
      resolve(chunk);
    });
  });
};

export { parseXml, parseJsonStream, parseFile };