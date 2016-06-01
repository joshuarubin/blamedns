import React from 'react';
import {render} from 'react-dom';
import { Router, Route, browserHistory, withRouter } from 'react-router'
import App from './app.jsx';
import Logs from './logs.jsx';
import NoMatch from './nomatch.jsx';

require("../vendor/skeleton/css/normalize.css");
require("../vendor/skeleton/css/skeleton.css");
require("../sass/style.scss");

render((
    <Router history={browserHistory}>
        <Route path="/ui"               component={App} />
        <Route path="/ui/logs"           component={withRouter(Logs)}/>
        <Route path="/ui/logs/:logLevel" component={withRouter(Logs)}/>
        <Route path="*"               component={NoMatch}/>
    </Router>
), document.getElementById('app'));
