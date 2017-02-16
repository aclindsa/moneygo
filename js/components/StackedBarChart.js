var d3 = require('d3');
var React = require('react');

var Panel = require('react-bootstrap').Panel;

module.exports = React.createClass({
	displayName: "StackedBarChart",
	calcMinMax: function(data) {
		var children = [];
		for (var child in data) {
			if (data.hasOwnProperty(child))
				children.push(data[child]);
		}

		var positiveValues = [0];
		var negativeValues = [0];
		if (children.length > 0 && children[0].length > 0) {
			for (var j = 0; j < children[0].length; j++) {
				positiveValues.push(children.reduce(function(accum, curr, i, arr) {
					if (arr[i][j] > 0)
						return accum + arr[i][j];
					return accum;
				}, 0));
				negativeValues.push(children.reduce(function(accum, curr, i, arr) {
					if (arr[i][j] < 0)
						return accum + arr[i][j];
					return accum;
				}, 0));
			}
		}

		return [Math.min.apply(Math, negativeValues), Math.max.apply(Math, positiveValues)];
	},
	calcAxisMarkSeparation: function(minMax, height, ticksPerHeight) {
		var targetTicks = height / ticksPerHeight;
		var range = minMax[1]-minMax[0];
		var rangePerTick = range/targetTicks;
		var roundOrder = Math.floor(Math.log(rangePerTick) / Math.LN10);
		var roundTo = Math.pow(10, roundOrder);
		return Math.ceil(rangePerTick/roundTo)*roundTo;
	},
	render: function() {
		var data = this.props.data.mapReduceChildren(null,
			function(accumulator, currentValue, currentIndex, array) {
				return accumulator + currentValue;
			}
		);

		height = 400;
		width = 600;
		legendWidth = 100;
		xMargin = 70;
		yMargin = 70;
		height -= yMargin*2;
		width -= xMargin*2;

		var minMax = this.calcMinMax(data);
		var y = d3.scaleLinear()
			.range([0, height])
			.domain(minMax);

		var xAxisMarksEvery = this.calcAxisMarkSeparation(minMax, height, 40);

		var x = d3.scaleLinear()
			.range([0, width])
			.domain([0, this.props.data.Labels.length + 0.5]);

		var bars = [];
		var labels = [];

		var barWidth = x(0.75);
		var barStart = x(0.25) + (x(1) - barWidth)/2;
		var childId=0;

		// Add Y axis marks and labels, and initialize positive- and
		// negativeSum arrays
		var positiveSum = [];
		var negativeSum = [];
		for (var i=0; i < this.props.data.Labels.length; i++) {
			positiveSum.push(0);
			negativeSum.push(0);
			var labelX = x(i) + barStart + barWidth/2;
			var labelY = height + 15;
			labels.push((
				<text x={labelX} y={labelY} transform={"rotate(45 "+labelX+" "+labelY+")"}>{this.props.data.Labels[i]}</text>
			));
			labels.push((
				<line className="axis-tick" x1={labelX} y1={height-3} x2={labelX} y2={height+3} />
			));
		}

		// Make X axis marks and labels
		var makeXLabel = function(value) {
			labels.push((
					<line className="axis-tick" x1={-3} y1={height - y(value)} x2={3} y2={height - y(value)} />
			));
			labels.push((
					<text is x={-10} y={height - y(value) + 6} text-anchor={"end"}>{value}</text>
			));
		}
		for (var i=0; i < minMax[1]; i+= xAxisMarksEvery)
			makeXLabel(i);
		for (var i=0-xAxisMarksEvery; i > minMax[0]; i -= xAxisMarksEvery)
			makeXLabel(i);

		//TODO handle Values from current series?
		var legendMap = {};
		for (var child in data) {
			childId++;
			var rectClasses = "chart-element chart-color" + (childId % 12);
			if (data.hasOwnProperty(child)) {
				for (var i=0; i < data[child].length; i++) {
					var value = data[child][i];
					if (value == 0)
						continue;
					legendMap[child] = childId;
					if (value > 0) {
						rectHeight = y(value) - y(0);
						positiveSum[i] += rectHeight;
						rectY = height - y(0) - positiveSum[i];
					} else {
						rectHeight = y(0) - y(value);
						rectY = height - y(0) + negativeSum[i];
						negativeSum[i] += rectHeight;
					}

					bars.push((
						<rect className={rectClasses} x={x(i) + barStart} y={rectY} width={barWidth} height={rectHeight} rx={1} ry={1}/>
					));
				}
			}
		}

		var legend = [];
		for (var series in legendMap) {
			var legendClasses = "chart-color" + (legendMap[series] % 12);
			var legendY = (legendMap[series] - 1)*15;
			legend.push((
				<rect className={legendClasses} x={0} y={legendY} width={10} height={10}/>
			));
			legend.push((
				<text x={0 + 15} y={legendY + 10}>{series}</text>
			));
		}

		return (
			<Panel header={this.props.data.Title}>
				<svg height={height + 2*yMargin} width={width + 2*xMargin + legendWidth}>
					<g className="stacked-bar-chart" transform={"translate("+xMargin+" "+yMargin+")"}>
						{bars}
						<line className="axis x-axis" x1={0} y1={height} x2={width} y2={height} />
						<line className="axis y-axis" x1={0} y1={0} x2={0} y2={height} />
						{labels}
					</g>
					<g className="chart-legend" transform={"translate("+(width + 2*xMargin)+" "+yMargin+")"}>
						{legend}
					</g>
				</svg>
			</Panel>
		);
	}
});
