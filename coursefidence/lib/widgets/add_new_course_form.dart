import 'package:coursefidence/models/course_model.dart';
import 'package:coursefidence/models/courseDS.dart';
import 'package:coursefidence/widgets/confidence_slide.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:flutter/services.dart';

class CourseForm extends StatefulWidget {
  @override
  CourseFormState createState() {
    return CourseFormState();
  }
}

class CourseFormState extends State<CourseForm> {
  final _formKey = GlobalKey<FormState>();

  final courseData = CourseDS();

  final focusCourseUnits = FocusNode();

  @override
  Widget build(BuildContext context) {
    // final courseModel = Provider.of<CourseModel>(context, listen: false);
    final courseModel = Provider.of<CourseModel>(context);

    return Form(
      key: _formKey,
      child: Container(
        padding: EdgeInsets.symmetric(horizontal: 40, vertical: 20),
        height: MediaQuery.of(context).size.height * 0.65,
        child: ListView(
          children: <Widget>[
            Text(
              "Add courses",
              style: TextStyle(fontSize: 25),
            ),
            TextFormField(
              decoration: InputDecoration(
                icon: Icon(Icons.book),
                hintText: 'Course Name',
                labelText: 'Name *',
              ),
              validator: (value) {
                if (value.isEmpty) {
                  return 'Please enter the course name';
                }
                return null;
              },
              onSaved: (String value) {
                courseData.courseName = value;
              },
              onFieldSubmitted: (v) {
                FocusScope.of(context).requestFocus(focusCourseUnits);
              },
            ),
            TextFormField(
              keyboardType: TextInputType.number,
              inputFormatters: [WhitelistingTextInputFormatter.digitsOnly],
              decoration: InputDecoration(
                icon: Icon(Icons.confirmation_number),
                hintText: 'Course Units',
                labelText: 'Units *',
              ),
              validator: (value) {
                if (value.isEmpty) {
                  return 'Please enter the course unit';
                }
                return null;
              },
              onSaved: (String value) {
                courseData.courseUnits = value;
              },
              focusNode: focusCourseUnits,
            ),
            SizedBox(height: 30),
            ConfidenceSlide(course: courseData),
            SizedBox(height: 30),
            Builder(
              builder: (BuildContext context) {
                return RaisedButton(
                  onPressed: () {
                    if (_formKey.currentState.validate()) {
                      _formKey.currentState.save();
                      courseModel.add(courseData);
                      Navigator.of(context).pop();
                    }
                  },
                  child: Text('Submit'),
                );
              },
            ),
          ],
        ),
      ),
    );
  }
}
