/**
* JetBrains Space Automation
* This Kotlin-script file lets you automate build activities
* For more info, see https://www.jetbrains.com/help/space/automation.html
*/

job("Hello World!") {
    container(displayName = "Say Hello", image = "hello-world")
}
job("Example shell script") {
    container(displayName = "Say Hello", image = "ubuntu") {
        shellScript {
            content = """
                echo Hello
                echo World!
            """
        }
    }
}