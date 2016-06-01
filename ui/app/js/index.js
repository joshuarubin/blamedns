import React from 'react'
import { render } from 'react-dom'
import { Router, Route, Redirect, browserHistory } from 'react-router'
import { createStore, combineReducers } from 'redux'
import { Provider } from 'react-redux'
import { syncHistoryWithStore, routerReducer } from 'react-router-redux'
import { App, LogsApp, NoMatch } from './components'
import * as reducers from './reducers'

require("../vendor/skeleton/css/normalize.css")
require("../vendor/skeleton/css/skeleton.css")
require("../sass/style.scss")

const store = createStore(
    combineReducers({
        ...reducers,
        routing: routerReducer
    })
)

const history = syncHistoryWithStore(browserHistory, store)

render((
    <Provider store={store}>
        <Router history={history}>
            <Route    path="/ui"                component={App}  />
            <Route    path="/ui/logs/:logLevel" component={LogsApp} />
            <Redirect from="/ui/logs" to="/ui/logs/warn" />
            <Route path="*"                  component={NoMatch} />
        </Router>
    </Provider>
), document.getElementById('app'))
