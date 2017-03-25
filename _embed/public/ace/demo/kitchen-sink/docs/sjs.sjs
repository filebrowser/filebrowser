var { each, map } = require('sjs:sequence');
var { get } = require('sjs:http');

function foo(items, nada) {
    var component = { name: "Ace", role: "Editor" };
    console.log("
        Welcome, #{component.name}
    ".trim());

    logging.debug(`Component added: $String(component) (${component})`);

    console.log(`
        Welcome, {${function() {
            return { x: 1, y: "why?}"};
        }()}
    `.trim());

    waitfor {
        items .. each.par { |item|
            get(item);
        }
    } and {
        var lengths = items .. map(i -> i.length);
    } or {
        hold(1500);
        throw new Error("timed out");
    }
}	// Real Tab.
