import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import org.kde.plasma.components as PlasmaComponents
import org.kde.plasma.core as PlasmaCore
import org.kde.plasma.plasmoid 2.0
import org.kde.kirigami as Kirigami

PlasmoidItem {
    id: root
    width: 300
    height: 220
    
    property string apiUrl: plasmoid.configuration.apiUrl || "http://localhost:8080"
    property bool isPlaying: false
    property bool isPaused: false
    property string trackTitle: ""
    property string trackArtist: ""
    property string trackThumbnail: ""
    property int volume: 50
    property string repeatMode: "off"
    property bool shuffle: false
    property bool connected: false
    property bool hasError: false
    property string errorMessage: ""
    property bool showSearchResults: false
    property var searchResults: []
    property var playlists: []
    property var queue: []
    property int currentQueuePosition: -1
    property int currentView: 0
    
    function get(url, endpoint) {
        var xhr = new XMLHttpRequest();
        xhr.open("GET", url + endpoint, false);
        xhr.setRequestHeader("Content-Type", "application/json");
        try {
            xhr.send();
            if (xhr.status === 200) {
                return JSON.parse(xhr.responseText);
            }
        } catch (e) {
            console.error("API error:", e);
        }
        return null;
    }
    
    function post(url, endpoint, data) {
        var xhr = new XMLHttpRequest();
        xhr.open("POST", url + endpoint, false);
        xhr.setRequestHeader("Content-Type", "application/json");
        try {
            xhr.send(JSON.stringify(data || {}));
            return xhr.status >= 200 && xhr.status < 300;
        } catch (e) {
            console.error("API error:", e);
            return false;
        }
    }
    
    function updatePlayerState() {
        var state = get(apiUrl, "/player/state");
        
        if (state) {
            connected = true;
            hasError = false;
            isPlaying = state.is_playing;
            isPaused = state.is_paused;
            volume = state.volume;
            repeatMode = state.repeat;
            shuffle = state.shuffle;
            
            if (state.current_track) {
                trackTitle = state.current_track.title;
                trackArtist = state.current_track.artist;
                trackThumbnail = state.current_track.thumbnail_url;
            } else {
                trackTitle = "";
                trackArtist = "";
                trackThumbnail = "";
            }
        } else {
            connected = false;
            hasError = true;
            errorMessage = "Cannot connect to API";
        }
        
        loadQueue();
    }
    
    function playPause() {
        post(apiUrl, "/player/pause", {});
        updatePlayerState();
    }
    
    function next() {
        post(apiUrl, "/player/next", {});
        updatePlayerState();
    }
    
    function previous() {
        post(apiUrl, "/player/previous", {});
        updatePlayerState();
    }
    
    function setVolume(vol) {
        post(apiUrl, "/player/volume", { volume: vol });
    }
    
    function toggleShuffle() {
        post(apiUrl, "/player/shuffle", {});
        updatePlayerState();
    }
    
    function cycleRepeat() {
        var state = get(apiUrl, "/player/state");
        if (!state) return;
        
        var nextRepeat = "off";
        switch (state.repeat) {
            case "off": nextRepeat = "all"; break;
            case "all": nextRepeat = "one"; break;
            case "one": nextRepeat = "off"; break;
        }
        
        post(apiUrl, "/player/repeat", { repeat: nextRepeat });
        updatePlayerState();
    }
    
    function search(query) {
        if (!query || query.trim() === "") {
            showSearchResults = false;
            searchResults = [];
            return;
        }
        var results = get(apiUrl, "/search?q=" + encodeURIComponent(query));
        if (results && results.results) {
            searchResults = results.results;
            showSearchResults = true;
        }
    }
    
    function playTrack(videoId) {
        post(apiUrl, "/player/play", { video_id: videoId });
        updatePlayerState();
        showSearchResults = false;
    }
    
    function playTrackNext(videoId) {
        var xhr = new XMLHttpRequest();
        xhr.open("POST", apiUrl + "/queue/play-next", false);
        xhr.setRequestHeader("Content-Type", "application/json");
        try {
            xhr.send(JSON.stringify({ video_id: videoId }));
            updatePlayerState();
            loadQueue();
        } catch (e) {
            console.error("API error:", e);
        }
        showSearchResults = false;
    }
    
    function loadQueue() {
        var result = get(apiUrl, "/queue");
        if (result && result.items) {
            queue = result.items;
            currentQueuePosition = result.current_position;
        }
    }
    
    function loadPlaylists() {
        var result = get(apiUrl, "/playlists");
        if (result && result.playlists) {
            playlists = result.playlists;
        }
    }
    
    function playPlaylist(playlistId) {
        post(apiUrl, "/playlists/" + playlistId + "/play", {});
        updatePlayerState();
    }
    
    Timer {
        id: timer
        interval: 2000
        running: true
        repeat: true
        onTriggered: updatePlayerState()
    }

    Rectangle {
        id: background
        anchors.fill: parent
        color: "transparent"
        
        ColumnLayout {
            anchors.fill: parent
            anchors.margins: 8
            spacing: 6
            
            RowLayout {
                Layout.fillWidth: true
                spacing: 6
                
                Rectangle {
                    width: 56
                    height: 56
                    radius: 4
                    color: Kirigami.Theme.backgroundColor
                    Layout.preferredWidth: 56
                    Layout.preferredHeight: 56
                    
                    Image {
                        id: thumbnail
                        source: trackThumbnail || ""
                        fillMode: Image.PreserveAspectCrop
                        anchors.fill: parent
                        visible: trackThumbnail !== ""
                    }
                    
                    Kirigami.Icon {
                        source: "../images/icon.png"
                        anchors.centerIn: parent
                        width: 24
                        height: 24
                        visible: trackThumbnail === ""
                    }
                }
                
                ColumnLayout {
                    Layout.fillWidth: true
                    spacing: 0
                    
                    PlasmaComponents.Label {
                        text: trackTitle || "Not Playing"
                        font.bold: true
                        font.pixelSize: 13
                        color: Kirigami.Theme.textColor
                        elide: Text.ElideRight
                        maximumLineCount: 1
                    }
                    
                    PlasmaComponents.Label {
                        text: trackArtist || "No track"
                        font.pixelSize: 11
                        color: Kirigami.Theme.disabledTextColor
                        elide: Text.ElideRight
                        maximumLineCount: 1
                    }
                }
            }
            
            RowLayout {
                Layout.fillWidth: true
                spacing: 4
                Layout.alignment: Qt.AlignHCenter
                
                PlasmaComponents.ToolButton {
                    icon.name: "media-skip-backward"
                    onClicked: previous()
                    enabled: connected
                    width: 32
                    height: 32
                }
                
                PlasmaComponents.ToolButton {
                    id: playBtn
                    icon.name: isPlaying ? "media-playback-pause" : "media-playback-start"
                    onClicked: playPause()
                    enabled: connected
                    width: 36
                    height: 36
                }
                
                PlasmaComponents.ToolButton {
                    icon.name: "media-skip-forward"
                    onClicked: next()
                    enabled: connected
                    width: 32
                    height: 32
                }
                
                PlasmaComponents.ToolButton {
                    icon.name: "view-refresh"
                    onClicked: toggleShuffle()
                    enabled: connected
                    checkable: true
                    checked: shuffle
                    width: 32
                    height: 32
                    opacity: shuffle ? 1.0 : 0.5
                }
                
                PlasmaComponents.ToolButton {
                    icon.name: repeatMode === "off" ? "media-repeat-none" : (repeatMode === "one" ? "media-repeat-one" : "media-repeat-all")
                    onClicked: cycleRepeat()
                    enabled: connected
                    width: 32
                    height: 32
                    opacity: repeatMode !== "off" ? 1.0 : 0.5
                }
            }
            
            RowLayout {
                Layout.fillWidth: true
                spacing: 6
                
                PlasmaComponents.Slider {
                    id: volumeSlider
                    Layout.fillWidth: true
                    from: 0
                    to: 100
                    value: volume
                    onMoved: setVolume(value)
                }
                
                PlasmaComponents.Label {
                    text: volume + "%"
                    font.pixelSize: 10
                    color: Kirigami.Theme.disabledTextColor
                    width: 35
                }
            }
            
            RowLayout {
                Layout.fillWidth: true
                spacing: 4
                
                PlasmaComponents.TextField {
                    id: searchField
                    Layout.fillWidth: true
                    placeholderText: "Search..."
                    onAccepted: search(text)
                    height: 28
                    font.pixelSize: 12
                }
            }
            
            RowLayout {
                Layout.fillWidth: true
                spacing: 4
                
                PlasmaComponents.Button {
                    text: "Now Playing"
                    onClicked: { currentView = 0; showSearchResults = false; }
                    checked: currentView === 0
                    checkable: true
                    height: 24
                    font.pixelSize: 10
                    Layout.fillWidth: true
                }
                
                PlasmaComponents.Button {
                    text: "Queue"
                    onClicked: { currentView = 1; loadQueue(); showSearchResults = false; }
                    checked: currentView === 1
                    checkable: true
                    height: 24
                    font.pixelSize: 10
                    Layout.fillWidth: true
                }
                
                PlasmaComponents.Button {
                    text: "Playlists"
                    onClicked: { currentView = 2; loadPlaylists(); showSearchResults = false; }
                    checked: currentView === 2
                    checkable: true
                    height: 24
                    font.pixelSize: 10
                    Layout.fillWidth: true
                }
            }
            
            Rectangle {
                Layout.fillWidth: true
                Layout.fillHeight: true
                color: Kirigami.Theme.backgroundColor
                radius: 4
                
                ScrollView {
                    anchors.fill: parent
                    visible: currentView === 0 && showSearchResults && searchResults.length > 0
                    
                    ColumnLayout {
                        width: parent.width
                        spacing: 2
                        
                        Repeater {
                            model: searchResults
                            delegate: MouseArea {
                                width: parent.width
                                height: 32
                                onClicked: {
                                    if (modelData.track) {
                                        playTrack(modelData.track.video_id);
                                    } else if (modelData.playlist) {
                                        playPlaylist(modelData.playlist.id);
                                    }
                                }
                                
                                RowLayout {
                                    anchors.fill: parent
                                    anchors.margins: 4
                                    spacing: 6
                                    
                                    Kirigami.Icon {
                                        source: modelData.track ? "audio-x-generic" : "folder"
                                        width: 20
                                        height: 20
                                    }
                                    
                                    PlasmaComponents.Label {
                                        text: modelData.track ? (modelData.track.title + " - " + modelData.track.artist) : 
                                               (modelData.playlist ? modelData.playlist.title : "")
                                        color: Kirigami.Theme.textColor
                                        font.pixelSize: 11
                                        elide: Text.ElideRight
                                        Layout.fillWidth: true
                                    }
                                    
                                    PlasmaComponents.ToolButton {
                                        icon.name: "go-next"
                                        onClicked: {
                                            if (modelData.track) {
                                                playTrackNext(modelData.track.video_id);
                                            }
                                        }
                                        width: 24
                                        height: 24
                                        opacity: 0.7
                                        visible: modelData.track
                                    }
                                }
                            }
                        }
                    }
                }
                
                ScrollView {
                    anchors.fill: parent
                    visible: currentView === 0 && (!showSearchResults || searchResults.length === 0)
                    
                    ColumnLayout {
                        width: parent.width
                        spacing: 2
                        
                        PlasmaComponents.Label {
                            text: "Up Next"
                            color: Kirigami.Theme.disabledTextColor
                            font.pixelSize: 10
                            font.bold: true
                            Layout.margins: 4
                        }
                        
                        Repeater {
                            model: queue
                            delegate: MouseArea {
                                width: parent.width
                                height: 36
                                onClicked: playTrack(modelData.track.video_id)
                                
                                RowLayout {
                                    anchors.fill: parent
                                    anchors.margins: 4
                                    spacing: 6
                                    
                                    PlasmaComponents.Label {
                                        text: modelData.position === currentQueuePosition ? ">" : (modelData.position + 1)
                                        color: modelData.position === currentQueuePosition ? Kirigami.Theme.highlightColor : Kirigami.Theme.disabledTextColor
                                        font.pixelSize: 10
                                        width: 20
                                    }
                                    
                                    Kirigami.Icon {
                                        source: "audio-x-generic"
                                        width: 20
                                        height: 20
                                    }
                                    
                                    ColumnLayout {
                                        Layout.fillWidth: true
                                        spacing: 0
                                        
                                        PlasmaComponents.Label {
                                            text: modelData.track.title
                                            color: modelData.position === currentQueuePosition ? Kirigami.Theme.highlightColor : Kirigami.Theme.textColor
                                            font.pixelSize: 11
                                            elide: Text.ElideRight
                                            Layout.fillWidth: true
                                        }
                                        
                                        PlasmaComponents.Label {
                                            text: modelData.track.artist
                                            color: Kirigami.Theme.disabledTextColor
                                            font.pixelSize: 10
                                            elide: Text.ElideRight
                                            Layout.fillWidth: true
                                        }
                                    }
                                    
                                    PlasmaComponents.ToolButton {
                                        icon.name: "list-add"
                                        onClicked: playTrackNext(modelData.track.video_id)
                                        width: 24
                                        height: 24
                                        opacity: 0.7
                                    }
                                }
                            }
                        }
                    }
                }
                
                ScrollView {
                    anchors.fill: parent
                    visible: currentView === 1
                    
                    ColumnLayout {
                        width: parent.width
                        spacing: 2
                        
                        Repeater {
                            model: queue
                            delegate: MouseArea {
                                width: parent.width
                                height: 40
                                onClicked: playTrack(modelData.track.video_id)
                                
                                RowLayout {
                                    anchors.fill: parent
                                    anchors.margins: 4
                                    spacing: 6
                                    
                                    PlasmaComponents.Label {
                                        text: modelData.position === currentQueuePosition ? ">" : (modelData.position + 1)
                                        color: modelData.position === currentQueuePosition ? Kirigami.Theme.highlightColor : Kirigami.Theme.disabledTextColor
                                        font.pixelSize: 10
                                        width: 20
                                    }
                                    
                                    Kirigami.Icon {
                                        source: "audio-x-generic"
                                        width: 20
                                        height: 20
                                    }
                                    
                                    ColumnLayout {
                                        Layout.fillWidth: true
                                        spacing: 0
                                        
                                        PlasmaComponents.Label {
                                            text: modelData.track.title
                                            color: modelData.position === currentQueuePosition ? Kirigami.Theme.highlightColor : Kirigami.Theme.textColor
                                            font.pixelSize: 11
                                            elide: Text.ElideRight
                                            Layout.fillWidth: true
                                        }
                                        
                                        PlasmaComponents.Label {
                                            text: modelData.track.artist
                                            color: Kirigami.Theme.disabledTextColor
                                            font.pixelSize: 10
                                            elide: Text.ElideRight
                                            Layout.fillWidth: true
                                        }
                                    }
                                    
                                    PlasmaComponents.ToolButton {
                                        icon.name: "list-add"
                                        onClicked: playTrackNext(modelData.track.video_id)
                                        width: 24
                                        height: 24
                                        opacity: 0.7
                                    }
                                }
                            }
                        }
                    }
                }
                
                ScrollView {
                    anchors.fill: parent
                    visible: currentView === 2
                    
                    ColumnLayout {
                        width: parent.width
                        spacing: 2
                        
                        Repeater {
                            model: playlists
                            delegate: MouseArea {
                                width: parent.width
                                height: 36
                                onClicked: playPlaylist(modelData.id)
                                
                                RowLayout {
                                    anchors.fill: parent
                                    anchors.margins: 4
                                    spacing: 6
                                    
                                    Kirigami.Icon {
                                        source: "folder"
                                        width: 24
                                        height: 24
                                    }
                                    
                                    ColumnLayout {
                                        Layout.fillWidth: true
                                        
                                        PlasmaComponents.Label {
                                            text: modelData.title
                                            color: Kirigami.Theme.textColor
                                            font.pixelSize: 12
                                            font.bold: true
                                            elide: Text.ElideRight
                                        }
                                        
                                        PlasmaComponents.Label {
                                            text: (modelData.track_count || 0) + " tracks"
                                            color: Kirigami.Theme.disabledTextColor
                                            font.pixelSize: 10
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}
