/* === This file is part of Calamares - <http://github.com/calamares> ===
 *
 * (Copyright and license details omitted for brevity)
 *
 * You should have received a copy of the GNU General Public License
 * along with Calamares. If not, see <http://www.gnu.org/licenses/>.
 */

import QtQuick 2.0;
import calamares.slideshow 1.0;

Presentation
{
    id: presentation

    function nextSlide() {
        console.log("QML Component (default slideshow) Next slide");
        presentation.goToNextSlide();
    }

    Timer {
        id: advanceTimer
        interval: 7500
        running: true
        repeat: true
        onTriggered: nextSlide()
    }

    // --- SLIDE 1 ---
    Slide {
        Image {
            id: reproductiveSystem
            source: "1-reproductive-system.png"
            anchors.centerIn: parent // CORRETTO: Centrato nel contenitore (Slide)
            width: 810
            height: 485
            fillMode: Image.PreserveAspectFit
        }
        Text {
            font.family: "Helvetica"
            font.pixelSize : 22
            color: "#aa3333"
            anchors.horizontalCenter: parent.horizontalCenter // CORRETTO: Centrato nel contenitore
            anchors.top: parent.top // CORRETTO: Ancorato in alto nel contenitore
            anchors.topMargin: 20   // Aggiunto un piccolo margine
            text: qsTr("<h1>Penguins' eggs</h1><br/>"+
                       "<h2>eggs: the reproductive system of penguins!</h2>"+
                       "<h3>https://penguins-eggs.net</h3>")
            wrapMode: Text.WordWrap
            width: 800
            horizontalAlignment: Text.Center
        }
    }

    // --- SLIDE 2 ---
    Slide {
        Image {
            id: startReproduction
            source: "2-start-reproduction.png"
            anchors.centerIn: parent // CORRETTO
            width: 810
            height: 485
            fillMode: Image.PreserveAspectFit
        }
        Text {
            font.family: "Helvetica"
            font.pixelSize : 22
            color: "#aa3333"
            anchors.horizontalCenter: parent.horizontalCenter // CORRETTO
            anchors.top: parent.top // CORRETTO
            anchors.topMargin: 20
            text: qsTr("<h1>Penguins' eggs</h1><br/>"+
                       "<h2>Start the reproduction of your system! Produce yours eggs and simply install them on another computer. You can do it!</h2>")
            wrapMode: Text.WordWrap
            width: 800
            horizontalAlignment: Text.Center
        }
    }

    // --- SLIDE 3 ---
    Slide {
        Image {
            id: itsYourSystem
            source: "3-its-your-system.png"
            anchors.centerIn: parent // CORRETTO
            width: 810
            height: 485
            fillMode: Image.PreserveAspectFit
        }
        Text {
            font.family: "Helvetica"
            font.pixelSize : 22
            color: "#aa3333"
            anchors.horizontalCenter: parent.horizontalCenter // CORRETTO
            anchors.top: parent.top // CORRETTO
            anchors.topMargin: 20
            text: qsTr("<h1>Penguins' eggs</h1><br/>"+
                       "<h2>It's your system!<br/>Bring it with you and use/install it elsewhere.</h2>")
            wrapMode: Text.WordWrap
            width: 800
            horizontalAlignment: Text.Center
        }
    }

    // --- SLIDE 4 ---
    Slide {
        Image {
            id: eggsPresentation
            source: "4-eggs-presentation.png"
            anchors.centerIn: parent // CORRETTO
            width: 810
            height: 485
            fillMode: Image.PreserveAspectFit
        }
        Text {
            font.family: "Helvetica"
            font.pixelSize : 22
            color: "#aa3333"
            anchors.horizontalCenter: parent.horizontalCenter // CORRETTO
            anchors.top: parent.top // CORRETTO
            anchors.topMargin: 20
            text: qsTr("<h1>Penguins' eggs</h1><br/>"+
                       "<h2>Eggs is a terminal tool,<br/>but it's simple and nice. Write eggs without else and get the commands list</h2>")
            wrapMode: Text.WordWrap
            width: 800
            horizontalAlignment: Text.Center
        }
    }

    // --- SLIDE 5 ---
    Slide {
        Image {
            id: waitHatching
            source: "5-wait-hatching.png"
            anchors.centerIn: parent // CORRETTO
            width: 810
            height: 485
            fillMode: Image.PreserveAspectFit
        }
        Text {
            font.family: "Helvetica"
            font.pixelSize : 22
            color: "#aa3333"
            anchors.horizontalCenter: parent.horizontalCenter // CORRETTO
            anchors.top: parent.top // CORRETTO
            anchors.topMargin: 20
            text: qsTr("<h1>Penguins' eggs</h1><br/>"+
                       "<h2>Please wait, we are hatching...<br/>Don't disturb the process, in few time You will have a new penguin.</h2>")
            wrapMode: Text.WordWrap
            width: 800
            horizontalAlignment: Text.Center
        }
    }

    // --- SLIDE 6 ---
    Slide {
        Image {
            id: followPenguins
            source: "6-follow-penguins.png"
            anchors.centerIn: parent // CORRETTO
            width: 810
            height: 485
            fillMode: Image.PreserveAspectFit
        }
        Text {
            font.family: "Helvetica"
            font.pixelSize : 22
            color: "#aa3333"
            anchors.horizontalCenter: parent.horizontalCenter // CORRETTO: Riferimento a 'slide6' rimosso
            anchors.top: parent.top // CORRETTO
            anchors.topMargin: 20
            text: qsTr("<h1>Penguins' eggs</h1><br/>"+
                       "<h3>Follow Penguins' eggs development, can be funny.</h3>" +
                       "<h3>Developers, graphics and testers need feedback and help.</h3>" +
                       "<h3>Use this tool, enjoy and if you can collaborate.</h3>"+
                       "<br><h2>That's all peoples!</h2>")
            wrapMode: Text.WordWrap
            width: 800
            horizontalAlignment: Text.Center
        }
    }

    // --- SLIDE 7 ---
    Slide {
        Image {
            id: createdBy
            source: "7-created-by.png"
            anchors.centerIn: parent // CORRETTO
            width: 810
            height: 485
            fillMode: Image.PreserveAspectFit
        }
        Text {
            font.family: "Helvetica"
            font.pixelSize : 22
            color: "#aa3333"
            anchors.horizontalCenter: parent.horizontalCenter // CORRETTO: Riferimento a 'followPenguins' rimosso
            anchors.top: parent.top // CORRETTO
            anchors.topMargin: 20
            text: qsTr("<h1>Penguins' eggs</h1><br/>"+
                       "<h2>Created by Piero Proietti</h2><br/>"+
                       "<h4>issues: htts://github.com/pieroproietti/penguins-eggs</h4>"+
                       "<h3>email: piero.proietti@gmail.com</h3>")
            wrapMode: Text.WordWrap
            width: 800
            horizontalAlignment: Text.Center
        }
    }
    
    function onActivate() {
        console.log("QML Component (default slideshow) activated");
        presentation.currentSlide = 0;
    }

    function onLeave() {
        console.log("QML Component (default slideshow) deactivated");
    }
}