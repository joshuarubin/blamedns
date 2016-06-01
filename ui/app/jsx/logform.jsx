import React from 'react';
import ReconnectingWebSocket from 'reconnectingwebsocket';

const relPathToAbs = (sRelPath) => {
    var nUpLn, sDir = "", sPath = location.pathname.replace(/[^\/]*$/, sRelPath.replace(/(\/|^)(?:\.?\/+)+/g, "$1"));
    for (var nEnd, nStart = 0; nEnd = sPath.indexOf("/../", nStart), nEnd > -1; nStart = nEnd + nUpLn) {
        nUpLn = /^\/(?:\.\.\/)*/.exec(sPath.slice(nEnd))[0].length;
        sDir = (sDir + sPath.substring(nStart, nEnd)).replace(new RegExp("(?:\\\/+[^\\\/]*){0," + ((nUpLn - 1) / 3) + "}$"), "/");
    }
    return sDir + sPath.substr(nStart);
}

export default React.createClass({
    propTypes: {
        initialEndpoint: React.PropTypes.string,
        initialLevel: React.PropTypes.string,
        onMessage: React.PropTypes.func.isRequired,
        onLevelChange: React.PropTypes.func
    },


    getDefaultProps: function() {
        return {
            initialEndpoint: "/v1/logs",
            initialLevel: ""
        };
    },

    getInitialState() {
        return {
            endpoint: this.props.initialEndpoint,
            level: this.props.initialLevel
        }
    },

    onSubmit(ev) {
        ev.preventDefault();
        this.open();
    },

    open() {
        this.close();

        var endpoint = this.endpoint()

        if (endpoint === "") {
            return;
        }

        this.ws = new ReconnectingWebSocket(endpoint);
        this.ws.onmessage = this.props.onMessage
    },

    close() {
        if (this.ws) {
            this.ws.close();
        }
    },

    onEndpointChange(ev) {
        this.setState({endpoint: ev.target.value});
    },

    onLevelChange(ev) {
        var level = ev.target.value;
        this.setState({level: level});
        if (this.props.onLevelChange) {
            this.props.onLevelChange(level);
        }
    },

    componentWillMount() {
        this.open();
    },

    componentWillUnmount() {
        this.close();
    },

    componentDidUpdate() {
        this.open();
    },

    endpoint() {
        var ep = this.state.endpoint;

        if (!ep || ep === "") {
            return ""
        }

        if (ep[0] === '.') {
            ep = relPathToAbs(ep);
        }

        ep = location.host + ep;

        return "ws://"+ep+"/"+this.state.level
    },

    render() {
        return (
            <form className="row" onSubmit={this.handleSubmit}>
                {/* <div className="seven columns"> */}
                {/*     <label htmlFor="logEndpoint">Log Endpoint</label> */}
                {/*     <input */}
                {/*         className="u-full-width" */}
                {/*         type="text" */}
                {/*         id="logEndpoint" */}
                {/*         value={this.state.endpoint} */}
                {/*         onChange={this.onEndpointChange} */}
                {/*     /> */}
                {/* </div> */}
                <div className="five columns">&nbsp;</div>
                <div className="two columns">
                    <label htmlFor="levelSelector">Log Level</label>
                    <select
                        className="u-full-width"
                        id="levelSelector"
                        value={this.state.level}
                        onChange={this.onLevelChange}
                    >
                        <option value="">Default</option>
                        <option value="debug">DEBUG</option>
                        <option value="info">INFO</option>
                        <option value="warn">WARN</option>
                        <option value="error">ERROR</option>
                        <option value="fatal">FATAL</option>
                        <option value="panic">PANIC</option>
                    </select>
                </div>
                {/* <input className="two columns button-primary u-full-width" type="submit" value="Submit" onClick={this.onSubmit} /> */}
            </form>
        );
    }
});
