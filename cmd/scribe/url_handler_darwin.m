#import <Cocoa/Cocoa.h>

// Forward declaration of the Go callback - CGO generates this
extern void urlCallbackGo(char* url);

// Objective-C class to handle Apple Events
@interface URLEventHandler : NSObject
+ (void)registerHandler;
@end

@implementation URLEventHandler
+ (void)registerHandler {
    NSAppleEventManager *em = [NSAppleEventManager sharedAppleEventManager];
    [em setEventHandler:self
            andSelector:@selector(handleGetURLEvent:withReplyEvent:)
          forEventClass:kInternetEventClass
             andEventID:kAEGetURL];
}

+ (void)handleGetURLEvent:(NSAppleEventDescriptor *)event withReplyEvent:(NSAppleEventDescriptor *)replyEvent {
    NSString *urlString = [[event paramDescriptorForKeyword:keyDirectObject] stringValue];
    if (urlString) {
        urlCallbackGo((char*)[urlString UTF8String]);
    }
}
@end

void RegisterURLHandler(void) {
    [URLEventHandler registerHandler];
}
