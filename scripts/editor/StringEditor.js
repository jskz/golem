/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
Golem.StringEditor = function (client, string, callback) {
    let _string = String(string);
    let _initialLineMatches = string.match(/[^\r\n]+/g);
    let cursor = _initialLineMatches ? _initialLineMatches.length : 0;

    const commit = () => (string = String(_string));
    const ch = client.character;

    function showEditorBuffer() {
        const lines = _string.match(/[^\r\n]+/g);

        if(lines) {
            for(let i = 0; i < lines.length; i++) {
                ch.send((i + 1) + ". " + lines[i] + "\r\n");
            }
        }
    }

    ch.send(
        "\r\n{Y! PRINT | @ DISCARD | $ WRITE | :# GOTO | :d# DELETE | :i# NEW LINES\r\n" +
        "-[ STRING EDITOR BUFFER ]---------------------------------------------{x\r\n");
    showEditorBuffer();
    ch.send("{Y----------------------------------------------------------------------{x\r\n");

    client.connectionHandler = function (input) {
        try {     
            if(input.length > Golem.StringEditor.MaxLineLength) {
                ch.send("{RInput too long, ignoring.{x\r\n");
                return;
            }

            if(!['!', '@', '$', '?'].includes(input) && input[0] !== ':') {
                const lines = _string.match(/[^\r\n]+/g);

                if(lines && lines.length + 1 > Golem.StringEditor.MaxAllowedLines) {
                    ch.send("{RString buffer is full, delete lines to make some room!{x\r\n");
                    return;
                }

                _string = lines
                    ? lines.filter((_, index) => index < cursor)
                        .concat([input])
                        .concat(lines.filter((_, index) => index > cursor))
                        .join("\r\n")
                    : input;
                cursor++;
            }

            if(input === '$') {
                commit();
                client.connectionHandler = null;

                callback(client, _string);
            } else if(input === '@') {
                client.connectionHandler = null;

                ch.send("{YExiting without writing changes.{x\r\n");
                callback(client, string);
            } else if(input === '!') {
                ch.send("{YContents of the string editor buffer:{x\r\n");

                showEditorBuffer();
            } else if(Golem.StringEditor.LineNavigationRegex.test(input)) {
                let [_, lineMatch] = Golem.StringEditor.LineNavigationRegex.exec(input);
                let m = null;
                let lineNumber = parseInt(lineMatch);
                
                const lines = [];

                while (null !== (m = Golem.StringEditor.LinesRegex.exec(_string))) {
                    lines.push(m.index);
                }

                if(lineNumber > lines.length) {
                    ch.send("{R! Cannot move out of bounds (" + lines.length + " lines in buffer){x\r\n");
                    return;
                }

                ch.send("{WOk.  Cursor at " + lineNumber + ". Next line will write to line " + (parseInt(lineNumber) + 1) + "{x\r\n");
                cursor = lineNumber;
            } else if(Golem.StringEditor.DeleteLineRegex.test(input)) {
                const [_, lineMatch] = Golem.StringEditor.DeleteLineRegex.exec(input),
                    lines = _string.match(/[^\r\n]+/g);
                const lineNumber = parseInt(lineMatch);

                if(lineNumber > Golem.StringEditor.MaxAllowedLines) return;
                if(lines) {
                    // lines in buffer prior to cursor
                    _string = lines.filter((_, index) => index < cursor)
                        // remaining lines after current lines + number being deleted
                        .concat(lines.filter(
                            (_, index) => (index >= cursor + lineNumber)))
                        // stringified
                        .join("\r\n");
                }

                ch.send("{WOk.  Deleted " + lineNumber + " lines from line " + lineNumber + "{x\r\n");
            } else if(Golem.StringEditor.InsertLineRegex.test(input)) {
                const [_, lineMatch] = Golem.StringEditor.InsertLineRegex.exec(input),
                    lines = _string.match(/[^\r\n]+/g);
                const lineNumber = parseInt(lineMatch);

                if(lines) {
                    if(lines.length + lineNumber > Golem.StringEditor.MaxAllowedLines) return;

                    // lines in buffer prior to the cursor
                    _string = lines.filter((_, index) => index < cursor)
                        // insert lineNumber empty lines
                        .concat(
                            [...Array(lineNumber)].map(_ => ("\r\n")))
                        // remainder of original string
                        .concat(lines.filter(
                            (_, index) => (index >= cursor)))
                        // stringified
                        .join("\r\n");
                } else {
                    _string = "\r\n".repeat(lineNumber);
                }

                ch.send("{WOk.  Inserted " + lineNumber + " empty lines at line " + cursor + "{x\r\n");
            }
        } catch(err) {
            ch.send(err);
        }
    };
};

Golem.StringEditor.MaxAllowedLines = 256;
Golem.StringEditor.MaxLineLength = 256;
Golem.StringEditor.LinesRegex = new RegExp(/[^\r\n]+/igm);
Golem.StringEditor.LineNavigationRegex = new RegExp(/^\:(\d+)/);
Golem.StringEditor.InsertLineRegex = new RegExp(/^\:i(\d+)/);
Golem.StringEditor.DeleteLineRegex = new RegExp(/^\:d(\d+)/);