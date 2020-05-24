import 'package:flutter/material.dart';
import 'package:flutter_staggered_grid_view/flutter_staggered_grid_view.dart';
import 'package:prescient_automation_flutter_test/IconGrid/TilesList.dart';

class IconGrid extends StatelessWidget {
  IconGrid({this.data});
  final data;


  @override
  Widget build(BuildContext context) {
  bool containsData = data[0].runtimeType == 'Text';

    return Padding(
      padding: const EdgeInsets.only(top: 12.0),
      child: StaggeredGridView.count(
        crossAxisCount: 4,
        staggeredTiles: staggeredTiles,
        children: dataTiles(data),
        mainAxisSpacing: 4.0,
        crossAxisSpacing: 4.0,
        padding: const EdgeInsets.all(4.0),
      ),
    );
  }
}
