const xsenv = require('@sap/xsenv');
const hana = require('@sap/hana-client');
var services = undefined;
var hanaConn = undefined;

module.exports.setupHana = function(){
    services = xsenv.getServices({
      hana: { label: 'hana' },
    });
    
    console.log("Creating connection to SAP hana db...")
    console.log(`host: ${services.hana.host}, port: ${services.hana.port}`)
    hanaConn = hana.createConnection();
}

module.exports.queryDB = async function(sql) {
    try {
      await hanaConn.connect(services.hana);
    } catch (err) {
      console.error('queryDB connect', err.message, err.stack);
      results = err.message;
    }
    try {
      console.log(`schema: ${services.hana.schema}`)
      await hanaConn.exec('SET SCHEMA ' + services.hana.schema);
      
      results = await hanaConn.exec(sql);
      
    } catch (err) {
      console.error('queryDB exec', err.message, err.stack);
      results = err.message;
    }
    try {
      await hanaConn.disconnect();
    } catch (err) {
      console.error('queryDB disconnect', err.message, err.stack);
      results = err.message;
    }
    return results;
}
