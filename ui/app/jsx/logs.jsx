import React from 'react';
import Log from './Log.jsx';
import LogForm from './LogForm.jsx';

export default React.createClass({
    propTypes: {
        router: React.PropTypes.shape({
            push: React.PropTypes.func.isRequired
        }).isRequired
    },

    getInitialState() {
        return {messages: [], logID: 0};
    },

    onMessage(ev) {
        this.setState({messages: this.state.messages.concat(JSON.parse(ev.data))});
    },

    onLevelChange(level) {
        var path = "/ui/logs";
        if (level && level.length > 0) {
            path += "/"+level;
        }

        this.props.router.push(path);
    },


    render() {
        var entries = [];
        for (var i = this.state.messages.length-1; i >=0; i--) {
            entries.push(this.state.messages[i]);
        }

        return (
            <div>
                <div className="log-form">
                    <LogForm onMessage={this.onMessage} initialLevel={this.props.params.logLevel} onLevelChange={this.onLevelChange} />
                </div>

                <div className="log-messages-container">
                    <div className="log-messages">
                        {entries.map(function(message, idx) {
                            return <Log key={idx} message={message} />;
                        })}
                    </div>
                </div>
            </div>
        );
    }
});
