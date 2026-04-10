import QtQuick
import QtQuick.Controls
import QtQuick.Layouts
import org.kde.kirigami as Kirigami
import org.kde.kcmutils as KCM

KCM.SimpleKCM {
    property alias cfg_apiUrl: apiUrlField.text

    Kirigami.FormLayout {
        TextField {
            id: apiUrlField
            Kirigami.FormData.label: "API URL:"
            placeholderText: "http://localhost:8080"
        }
    }
}
