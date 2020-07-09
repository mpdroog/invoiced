// Small bunch of prototype functions to rapidly
// play with data in memory
//

// Convert object to array and return subitems as one big array
Object.prototype.concatArray = function() {
	var n = Object.values(this);
	var out = [];
	for (i = 0; i < n.length; i++) {
		out = out.concat(n[i]);
	}
	return out;
}

Array.prototype.filterField = function(fieldname) {
	var out = [];
	var keys = fieldname.split(".");
	for (var i = 0; i < this.length; i++) {
		var base = this[i];
		for (var n = 0; n < keys.length; n++) {
			base = base[ keys[n] ];
		}
		out.push(base);
	}
	return out;
}
Array.prototype.remove = function(id) {
	// ignore res
	this.splice(id, 1);
	return this;
}

Array.prototype.sum = function(round) {
	var out = new Big("0.00");

	for (var i = 0; i < this.length; i++) {
		out = out.plus(new Big(this[i]));
	}

	return out.round(round).toFixed(round).toString();
}

console.log(`
Total ex VAT this year:
window.rootdev.invoiced.paid.concatArray().filterField("Total.Ex").sum(2);
Get worked hours this year:
window.rootdev.invoiced.paid.concatArray().filterField("Lines").concatArray().filterField("Quantity")`);