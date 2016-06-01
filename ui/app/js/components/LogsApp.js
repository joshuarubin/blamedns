import React, { Component, PropTypes } from 'react'
import { connect } from 'react-redux'
import { withRouter } from 'react-router'
import ReconnectingWebSocket from 'reconnectingwebsocket'
import LogForm from "./LogForm"
import Logs from "./Logs"
import { levelToString, stringToLevel } from "../lib"
import { addLogEntry } from '../actions'

class LogsApp extends Component {
    open() {
        this.close()

        let endpoint = "ws://" + location.host + "/v1/logs/" + this.props.params.logLevel

        this.ws = new ReconnectingWebSocket(endpoint)
        this.ws.onmessage = data => {
            this.props.onMessage(JSON.parse(data.data))
        }
    }

    close() {
        if (this.ws) {
            this.ws.close()
        }
    }

    componentWillMount() {
        this.open()
    }

    componentWillUnmount() {
        this.close()
    }

    componentDidUpdate() {
        this.open()
    }

    onLevelChange(ev) {
        let logLevel = parseInt(ev.target.value)
        let path = "/ui/logs/" + levelToString(logLevel).toLowerCase()

        this.props.router.push(path)
    }

    render() {
        let logLevel = stringToLevel(this.props.params.logLevel)

        return (
            <div>
                <div className="log-form">
                    <LogForm logLevel={logLevel} onLevelChange={this.onLevelChange.bind(this)} />
                </div>

                <div className="log-messages-container">
                    <div className="log-messages">
                        <Logs logLevel={logLevel} />
                    </div>
                </div>
            </div>
        );
    }
}

LogsApp.propTypes = {
    router: PropTypes.shape({
        push: PropTypes.func.isRequired
    }).isRequired,
    onMessage: PropTypes.func.isRequired
}

const mapDispatchToProps = (dispatch) => {
    return {
        onMessage: (data) => {
            dispatch(addLogEntry(data))
        }
    }
}

export default connect(
    undefined,
    mapDispatchToProps
)(withRouter(LogsApp))
