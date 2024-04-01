module.exports['without-ext'] = function (str) {
    return str.replace('.md', '');
}

module.exports['get-first-natspec'] = function (str, natspecProperty) {
    // Extract data from first natspec property that matches natspecProperty
    var match = str.match(new RegExp(`@${natspecProperty} ([^@]*)`));
    return match[1];
}