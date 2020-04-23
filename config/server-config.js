// Create config variables
var environments = {};

environments.dev = {
    'httpPort' : 3000,
    'metaLogging': false,
    'envName'  : 'dev',
    'apiKey'   : 'dev!env123'
};

// Check if environment variable for NODE_ENV was set
var currentEnvironment = typeof(process.env.NODE_ENV) == 'string' ? process.env.NODE_ENV.toLowerCase() : '';

// Check if environment set in NODE_ENV exists in envoronments object. If not, set it do dev
var environmenToExport = typeof(environments[currentEnvironment]) == 'object' ? environments[currentEnvironment] : environments.dev;

//Exoirt the module
module.exports = environmenToExport;