const express = require('express');
const router = express.Router();

// Middleware that is getting called on all other methosts and retursn HTTP 405
const methodNotAllowed = (req, res, next) => res.status(405).set('Content-Type', 'application/json').send(`{ 'success': false, 'explanation': 'method_not_allowed' }`);

const test = [
    {
        "alertName" : "testing",
        "provider"  : "banaana"
    }
];

router.get('/', (req, res) => {
	res.status(200).set('Content-Type', 'application/json').send(`{ 'success': true, explanation': 'alert_definitoon_saved' }`);
});

router.post('/', (req, res) => {
    if (!(req.is('application/json'))){
        res.status(415).set('Content-Type', 'application/json').send(`{ 'success': false, 'explanation': 'wrong_content_type' }`);
    }
    // Go to ES and check if exists


    if (test.some (test => test.alertName === req.body.alertName)){
        res.status(400).set('Content-Type', 'application/json').send(`{ 'success': false, 'explanation': 'alert_structure_exists' }`);
    }
    

    
    res.status(200).set('Content-Type', 'application/json').send(`{ 'success': true, explanation': 'alert_definitoon_saved' }`);

});

router.all('/', methodNotAllowed)

module.exports = router;