/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
Golem.StringEditor = function (client, string, callback) {
    let _string = string;
    let cursor = string.match(/[^\r\n]+/g).length;

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
        "\r\n{Y?SINGLE ! PRINT | $ WRITE/QUIT | :5 (line) SET LINE, :d # DELETE LINES\r\n" + 
        "-[ STRING EDITOR BUFFER ]---------------------------------------------{x\r\n");
    showEditorBuffer();
    ch.send("{Y----------------------------------------------------------------------{x\r\n");

    client.connectionHandler = function (input) {
        try {
            if(!['!', '$', '?'].includes(input) && input[0] !== ':') {
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
            } else if(input === '!') {
                ch.send("{YContents of the string editor buffer:{x\r\n");

                showEditorBuffer();
            } else if(Golem.StringEditor.LineNavigationRegex.test(input)) {
                let [_, lineNumber] = Golem.StringEditor.LineNavigationRegex.exec(input);
                let m = null;
                
                const lines = [];

                while (null !== (m = Golem.StringEditor.LinesRegex.exec(_string))) {
                    lines.push(m.index);
                }

                if(lineNumber > lines.length) {
                    ch.send("{R! Cannot move out of bounds (" + lines.length + " lines in buffer){x\r\n");
                    return;
                }

                ch.send("{WOk.  Moved to line " + lineNumber + ".{x\r\n");
                cursor = lineNumber;
            } else if(Golem.StringEditor.DeleteLineRegex.test(input)) {
                const [_, lineNumber] = Golem.StringEditor.DeleteLineRegex.exec(input),
                    lines = _string.match(/[^\r\n]+/g);

                if(lineNumber > 1024) return;
                if(lines) {
                    _string = lines.filter((_, index) => index < cursor)
                        .concat(lines.filter(
                            (_, index) => (index >= cursor + lineNumber)))
                        .join("\r\n");
                }

                ch.send("{WOk.  Deleted " + lineNumber + " lines.{x\r\n");
            }
        } catch(err) {
            ch.send(err);
        }
    };
};

Golem.StringEditor.MaxAllowedLines = 256;
Golem.StringEditor.LinesRegex = new RegExp(/[^\r\n]+/igm);
Golem.StringEditor.LineNavigationRegex = new RegExp(/^\:(\d+)/);
Golem.StringEditor.DeleteLineRegex = new RegExp(/^\:d(\d+)/);