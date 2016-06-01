import React, { PropTypes, Component } from 'react'
import moment from 'moment'
import { levelToString } from "../lib"

class Log extends Component {
    render() {
        let msg = this.props.entry
        let data = []
        for (let key in msg.Data) {
            if (msg.Data.hasOwnProperty(key)) {
                data.push({key: key, value: msg.Data[key]})
            }
        }

        return (
            <div className={"row log-entry log-level-"+msg.Level}>
                <div className="two columns log-time">
                    {moment(msg.Time).format("MMM D h:mm:ssa")}
                </div>
                <div className={"one column log-level log-level-"+msg.Level}>
                    {levelToString(msg.Level)}
                </div>
                <div className="two columns log-message">
                    {msg.Message}
                </div>
                <div className="seven columns log-data">
                    {data.map(function(d, idx) {
                        return (
                            <span key={idx}>
                                <span className="log-data-key">{d.key}</span>
                                <span className="log-data-value">{d.value}</span>
                            </span>
                        )
                    })}
                </div>
            </div>
        )
    }
}

Log.propTypes = {
    entry: PropTypes.shape({
        Data: PropTypes.object,
        Level: PropTypes.number.isRequired,
        Time: PropTypes.string.isRequired,
        Message: PropTypes.string.isRequired
    }).isRequired
}

export default Log
