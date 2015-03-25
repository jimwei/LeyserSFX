SET TargetFile=SFX\CS_QuickDeployment.exe
SET ToolPath=C:\_LeySerBuildTools\AutoFileSign

"%ToolPath%\signtool.exe" sign /f "%ToolPath%\LeySer.pfx" /p LeySerNo1 /t http://timestamp.verisign.com/scripts/timstamp.dll /d "LeySer System 8.5" "%TargetFile%"

PAUSE
